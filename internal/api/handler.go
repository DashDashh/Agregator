package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/internal/models"
	"github.com/kirilltahmazidi/aggregator/internal/store"
)

// Publisher — интерфейс для отправки сообщений в Kafka.
type Publisher interface {
	PublishOrder(ctx context.Context, order *store.Order) error
	PublishConfirmPrice(ctx context.Context, payload models.ConfirmPricePayload) error
}

// HTTP-обработчики REST API для фронтенда
type Handler struct {
	store     *store.Store
	publisher Publisher // отправляет заказы эксплуатантам через Kafka
}

func NewHandler(s *store.Store, p Publisher) *Handler {
	return &Handler{store: s, publisher: p}
}

// проверка что сервис жив
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /orders — создать новый заказ
//
// Тело запроса:
//
//		{
//		  "customer_id": "uuid",
//		  "description": "доставить посылку",
//		  "budget": 1500.0,
//		  "from_lat": 55.75,
//		  "from_lon": 37.61,
//		  "to_lat":   55.80,
//	   "to_lon":   37.65
//		}
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID  string  `json:"customer_id"`
		Description string  `json:"description"`
		Budget      float64 `json:"budget"`
		FromLat     float64 `json:"from_lat"`
		FromLon     float64 `json:"from_lon"`
		ToLat       float64 `json:"to_lat"`
		ToLon       float64 `json:"to_lon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.CustomerID == "" || req.Description == "" {
		respondError(w, http.StatusBadRequest, "customer_id и description обязательны")
		return
	}

	order := &store.Order{
		ID:          uuid.NewString(),
		CustomerID:  req.CustomerID,
		Description: req.Description,
		Budget:      req.Budget,
		FromLat:     req.FromLat,
		FromLon:     req.FromLon,
		ToLat:       req.ToLat,
		ToLon:       req.ToLon,
		Status:      store.StatusPending,
		CreatedAt:   time.Now(),
	}
	if err := h.store.SaveOrder(order); err != nil {
		log.Printf("[api] failed to save order: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка сохранения заказа")
		return
	}
	log.Printf("[api] order created id=%s customer=%s", order.ID, order.CustomerID)

	// Отправляем заказ эксплуатантам через Kafka — они его прочитают из operator.requests
	if err := h.publisher.PublishOrder(r.Context(), order); err != nil {
		log.Printf("[api] failed to publish order to kafka: %v", err)
		// не падаем — заказ уже сохранён, оператор получит его позже
	}

	respond(w, http.StatusCreated, order)
}

// GET /orders — список всех заказов
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders := h.store.ListOrders()
	respond(w, http.StatusOK, orders)
}

// GET /orders/{id} — статус конкретного заказа
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/orders/")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id заказа не указан")
		return
	}
	order, ok := h.store.GetOrder(id)
	if !ok {
		respondError(w, http.StatusNotFound, "заказ не найден")
		return
	}
	respond(w, http.StatusOK, order)
}

// POST /operators — зарегистрировать эксплуатанта
//
// Тело запроса:
//
//	{ "name": "ООО Дроны", "license": "LIC-001", "email": "ops@example.com" }
func (h *Handler) RegisterOperator(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		License string `json:"license"`
		Email   string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.Name == "" || req.License == "" {
		respondError(w, http.StatusBadRequest, "name и license обязательны")
		return
	}

	op := &store.Operator{
		ID:      uuid.NewString(),
		Name:    req.Name,
		License: req.License,
		Email:   req.Email,
	}
	if err := h.store.SaveOperator(op); err != nil {
		log.Printf("[api] failed to save operator: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	log.Printf("[api] operator registered id=%s name=%s", op.ID, op.Name)

	respond(w, http.StatusCreated, op)
}

// POST /customers — зарегистрировать заказчика
//
// Тело запроса:
//
//	{ "name": "Иван Иванов", "email": "ivan@example.com", "phone": "+79001234567" }
func (h *Handler) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.Name == "" || req.Email == "" {
		respondError(w, http.StatusBadRequest, "name и email обязательны")
		return
	}

	c := &store.Customer{
		ID:    uuid.NewString(),
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}
	if err := h.store.SaveCustomer(c); err != nil {
		log.Printf("[api] failed to save customer: %v", err)
		respondError(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	log.Printf("[api] customer registered id=%s name=%s", c.ID, c.Name)

	respond(w, http.StatusCreated, c)
}

// POST /orders/{id}/confirm-price — пользователь подтверждает цену эксплуатанта.
// Агрегатор отправляет сообщение confirm_price эксплуатанту через Kafka (operator.requests)
// и обновляет статус заказа на "confirmed".
//
// Тело запроса: { "operator_id": "uuid", "accepted_price": 4500.00 }
func (h *Handler) ConfirmPrice(w http.ResponseWriter, r *http.Request) {
	// извлекаем order_id из URL /orders/{id}/confirm-price
	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	orderID := strings.TrimSuffix(path, "/confirm-price")
	if orderID == "" {
		respondError(w, http.StatusBadRequest, "id заказа не указан")
		return
	}

	var req struct {
		OperatorID    string  `json:"operator_id"`
		AcceptedPrice float64 `json:"accepted_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.OperatorID == "" || req.AcceptedPrice <= 0 {
		respondError(w, http.StatusBadRequest, "operator_id и accepted_price обязательны")
		return
	}

	_, ok := h.store.GetOrder(orderID)
	if !ok {
		respondError(w, http.StatusNotFound, "заказ не найден")
		return
	}

	// Сменяем статус на confirmed и отправляем цену эксплуатанту
	h.store.UpdateOrderStatus(orderID, store.StatusConfirmed)

	payload := models.ConfirmPricePayload{
		OrderID:       orderID,
		OperatorID:    req.OperatorID,
		AcceptedPrice: req.AcceptedPrice,
	}
	if err := h.publisher.PublishConfirmPrice(r.Context(), payload); err != nil {
		log.Printf("[api] failed to publish confirm_price: %v", err)
		// не падаем — статус уже обновлён
	}
	log.Printf("[api] price confirmed order_id=%s operator=%s price=%.2f", orderID, req.OperatorID, req.AcceptedPrice)

	respond(w, http.StatusOK, map[string]interface{}{
		"order_id":       orderID,
		"operator_id":    req.OperatorID,
		"accepted_price": req.AcceptedPrice,
		"status":         "confirmed",
	})
}

// вспомогательные функции
func respond(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, code int, msg string) {
	respond(w, code, map[string]string{"error": msg})
}
