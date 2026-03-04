package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/internal/store"
)

// HTTP-обработчики REST API для фронтенда
type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
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
	h.store.SaveOrder(order)
	log.Printf("[api] order created id=%s customer=%s", order.ID, order.CustomerID)

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
	h.store.SaveOperator(op)
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
	h.store.SaveCustomer(c)
	log.Printf("[api] customer registered id=%s name=%s", c.ID, c.Name)

	respond(w, http.StatusCreated, c)
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
