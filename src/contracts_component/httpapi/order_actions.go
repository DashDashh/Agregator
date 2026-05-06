package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

// ConfirmPrice handles POST /orders/{id}/confirm-price.
func (h *Handler) ConfirmPrice(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if user.Role != "customer" {
		httpx.RespondError(w, http.StatusForbidden, "подтвердить цену может только заказчик")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	orderID := strings.TrimSuffix(path, "/confirm-price")
	if orderID == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id заказа не указан")
		return
	}

	var req struct {
		OperatorID    string  `json:"operator_id"`
		AcceptedPrice float64 `json:"accepted_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.OperatorID == "" || req.AcceptedPrice <= 0 {
		httpx.RespondError(w, http.StatusBadRequest, "operator_id и accepted_price обязательны")
		return
	}

	order, ok := h.store.GetOrder(orderID)
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, "заказ не найден")
		return
	}
	if order.CustomerID != user.ID {
		httpx.RespondError(w, http.StatusForbidden, "нельзя подтверждать чужой заказ")
		return
	}
	commission := req.AcceptedPrice * h.commissionRate
	operatorAmount := req.AcceptedPrice - commission
	if !h.store.ConfirmPrice(orderID, req.OperatorID, req.AcceptedPrice, commission) {
		httpx.RespondError(w, http.StatusBadRequest, "недопустимое состояние заказа или неверный оператор/цена")
		return
	}

	payload := models.ConfirmPricePayload{
		OrderID:          orderID,
		OperatorID:       req.OperatorID,
		AcceptedPrice:    req.AcceptedPrice,
		CommissionAmount: commission,
		OperatorAmount:   operatorAmount,
	}
	if err := h.publisher.PublishConfirmPrice(r.Context(), payload); err != nil {
		log.Printf("[api] failed to publish confirm_price: %v", err)
	}
	log.Printf("[api] price confirmed order_id=%s operator=%s price=%.2f", orderID, req.OperatorID, req.AcceptedPrice)

	httpx.Respond(w, http.StatusOK, map[string]interface{}{
		"order_id":          orderID,
		"operator_id":       req.OperatorID,
		"accepted_price":    req.AcceptedPrice,
		"commission_amount": commission,
		"operator_amount":   operatorAmount,
		"status":            "confirmed",
	})
}

// OfferPrice handles POST /orders/{id}/offer.
func (h *Handler) OfferPrice(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if user.Role != "operator" {
		httpx.RespondError(w, http.StatusForbidden, "предложить цену может только эксплуатант")
		return
	}

	orderID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/orders/"), "/offer")
	if orderID == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id заказа не указан")
		return
	}

	var req struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.Price <= 0 {
		httpx.RespondError(w, http.StatusBadRequest, "price должен быть больше 0")
		return
	}

	if _, ok := h.store.GetOrder(orderID); !ok {
		httpx.RespondError(w, http.StatusNotFound, "заказ не найден")
		return
	}
	if !h.store.SetOperatorOffer(orderID, user.ID, req.Price) {
		httpx.RespondError(w, http.StatusBadRequest, "нельзя предложить цену для текущего статуса заказа")
		return
	}

	httpx.Respond(w, http.StatusOK, map[string]interface{}{
		"order_id":      orderID,
		"operator_id":   user.ID,
		"offered_price": req.Price,
		"status":        "matched",
	})
}

// ConfirmCompletion handles POST /orders/{id}/confirm-completion.
func (h *Handler) ConfirmCompletion(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if user.Role != "customer" {
		httpx.RespondError(w, http.StatusForbidden, "подтвердить выполнение может только заказчик")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	orderID := strings.TrimSuffix(path, "/confirm-completion")
	if orderID == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id заказа не указан")
		return
	}

	order, ok := h.store.GetOrder(orderID)
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, "заказ не найден")
		return
	}
	if order.CustomerID != user.ID {
		httpx.RespondError(w, http.StatusForbidden, "нельзя подтверждать чужой заказ")
		return
	}

	if !h.store.ConfirmCompletion(orderID) {
		httpx.RespondError(w, http.StatusBadRequest, "недопустимое состояние: заказ еще не выполнен оператором")
		return
	}

	httpx.Respond(w, http.StatusOK, map[string]interface{}{
		"order_id": orderID,
		"status":   "completed",
	})
}
