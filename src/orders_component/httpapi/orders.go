package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

// CreateOrder handles POST /orders.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if user.Role != "customer" {
		httpx.RespondError(w, http.StatusForbidden, "создавать заказ может только заказчик")
		return
	}

	var req struct {
		CustomerID     string   `json:"customer_id"`
		Description    string   `json:"description"`
		Budget         float64  `json:"budget"`
		FromLat        float64  `json:"from_lat"`
		FromLon        float64  `json:"from_lon"`
		ToLat          float64  `json:"to_lat"`
		ToLon          float64  `json:"to_lon"`
		MissionType    string   `json:"mission_type"`
		SecurityGoals  []string `json:"security_goals"`
		TopLeftLat     float64  `json:"top_left_lat"`
		TopLeftLon     float64  `json:"top_left_lon"`
		BottomRightLat float64  `json:"bottom_right_lat"`
		BottomRightLon float64  `json:"bottom_right_lon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.Description == "" {
		httpx.RespondError(w, http.StatusBadRequest, "description обязателен")
		return
	}
	req.CustomerID = user.ID
	if _, ok := h.store.GetCustomer(req.CustomerID); !ok {
		httpx.RespondError(w, http.StatusNotFound, "заказчик не найден")
		return
	}
	missionType := req.MissionType
	if missionType == "" {
		missionType = "delivery"
	}

	order := &store.Order{
		ID:             uuid.NewString(),
		CustomerID:     req.CustomerID,
		Description:    req.Description,
		Budget:         req.Budget,
		FromLat:        req.FromLat,
		FromLon:        req.FromLon,
		ToLat:          req.ToLat,
		ToLon:          req.ToLon,
		MissionType:    missionType,
		SecurityGoals:  req.SecurityGoals,
		TopLeftLat:     req.TopLeftLat,
		TopLeftLon:     req.TopLeftLon,
		BottomRightLat: req.BottomRightLat,
		BottomRightLon: req.BottomRightLon,
		Status:         store.StatusPending,
		CreatedAt:      time.Now(),
	}
	if err := h.store.SaveOrder(order); err != nil {
		log.Printf("[api] failed to save order: %v", err)
		httpx.RespondError(w, http.StatusInternalServerError, "ошибка сохранения заказа")
		return
	}
	log.Printf("[api] order created id=%s customer=%s", order.ID, order.CustomerID)

	if err := h.publisher.PublishOrder(r.Context(), order); err != nil {
		log.Printf("[api] failed to publish order to kafka: %v", err)
	} else if ok := h.store.UpdateOrderStatus(order.ID, store.StatusSearching); !ok {
		log.Printf("[api] failed to update order status to searching: order_id=%s", order.ID)
	} else {
		order.Status = store.StatusSearching
	}

	httpx.Respond(w, http.StatusCreated, order)
}

// ListOrders handles GET /orders.
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if user.Role == "customer" {
		customerID := strings.TrimSpace(r.URL.Query().Get("customer_id"))
		if customerID != "" && customerID != user.ID {
			httpx.RespondError(w, http.StatusForbidden, "нельзя смотреть заказы другого заказчика")
			return
		}
		httpx.Respond(w, http.StatusOK, h.store.ListOrdersByCustomer(user.ID))
		return
	}

	if customerID := strings.TrimSpace(r.URL.Query().Get("customer_id")); customerID != "" {
		if _, ok := h.store.GetCustomer(customerID); !ok {
			httpx.RespondError(w, http.StatusNotFound, "заказчик не найден")
			return
		}
		httpx.Respond(w, http.StatusOK, h.store.ListOrdersByCustomer(customerID))
		return
	}

	orders := h.store.ListOrders()
	httpx.Respond(w, http.StatusOK, orders)
}

// GetOrder handles GET /orders/{id}.
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/orders/")
	if id == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id заказа не указан")
		return
	}
	order, ok := h.store.GetOrder(id)
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, "заказ не найден")
		return
	}
	if user.Role == "customer" && order.CustomerID != user.ID {
		httpx.RespondError(w, http.StatusForbidden, "нельзя смотреть чужой заказ")
		return
	}
	httpx.Respond(w, http.StatusOK, order)
}
