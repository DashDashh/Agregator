package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

// ConfirmPrice обрабатывает POST /orders/{id}/confirm-price.
func (h *Handler) ConfirmPrice(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.authRequired && user.Role != "customer" {
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
	if h.authRequired && order.CustomerID != user.ID {
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

// OfferPrice обрабатывает POST /orders/{id}/offer.
func (h *Handler) OfferPrice(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.authRequired && user.Role != "operator" {
		httpx.RespondError(w, http.StatusForbidden, "предложить цену может только эксплуатант")
		return
	}

	orderID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/orders/"), "/offer")
	if orderID == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id заказа не указан")
		return
	}

	var req struct {
		Price      float64 `json:"price"`
		OperatorID string  `json:"operator_id"`
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
	operatorID := user.ID
	if operatorID == "" {
		operatorID = strings.TrimSpace(req.OperatorID)
	}
	if operatorID == "" {
		operatorID = "integration-operator"
	}
	if !h.store.SetOperatorOffer(orderID, operatorID, req.Price) {
		httpx.RespondError(w, http.StatusBadRequest, "нельзя предложить цену для текущего статуса заказа")
		return
	}

	httpx.Respond(w, http.StatusOK, map[string]interface{}{
		"order_id":      orderID,
		"operator_id":   operatorID,
		"offered_price": req.Price,
		"status":        "matched",
	})
}

// ConfirmCompletion обрабатывает POST /orders/{id}/confirm-completion.
func (h *Handler) ConfirmCompletion(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.authRequired && user.Role != "customer" {
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
	if h.authRequired && order.CustomerID != user.ID {
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

// ReportIncident обрабатывает POST /orders/{id}/incident.
func (h *Handler) ReportIncident(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}
	if h.authRequired && user.Role != "customer" && user.Role != "operator" {
		httpx.RespondError(w, http.StatusForbidden, "сообщить об инциденте может заказчик или эксплуатант")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	orderID := strings.TrimSuffix(path, "/incident")
	if orderID == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id заказа не указан")
		return
	}

	var req models.IncidentReportedPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.OrderID != "" && req.OrderID != orderID {
		httpx.RespondError(w, http.StatusBadRequest, "order_id в теле не совпадает с id в пути")
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		httpx.RespondError(w, http.StatusBadRequest, "reason обязателен")
		return
	}
	if req.DamageAmount < 0 {
		httpx.RespondError(w, http.StatusBadRequest, "damage_amount не может быть отрицательным")
		return
	}

	order, found := h.store.GetOrder(orderID)
	if !found {
		httpx.RespondError(w, http.StatusNotFound, "заказ не найден")
		return
	}
	if h.authRequired && user.Role == "customer" && order.CustomerID != user.ID {
		httpx.RespondError(w, http.StatusForbidden, "нельзя создать инцидент по чужому заказу")
		return
	}
	if h.authRequired && user.Role == "operator" && order.OperatorID != "" && order.OperatorID != user.ID {
		httpx.RespondError(w, http.StatusForbidden, "нельзя создать инцидент по заказу другого эксплуатанта")
		return
	}

	operatorID := req.OperatorID
	if operatorID == "" {
		operatorID = order.OperatorID
	}
	incident := &domain.Incident{
		ID:           uuid.NewString(),
		OrderID:      orderID,
		OperatorID:   operatorID,
		ReporterID:   firstNonEmpty(user.ID, req.ReporterID, order.CustomerID),
		Reason:       strings.TrimSpace(req.Reason),
		Description:  strings.TrimSpace(req.Description),
		DamageAmount: req.DamageAmount,
		Status:       "registered",
		CreatedAt:    time.Now().UTC(),
	}
	if err := h.store.RegisterIncident(incident); err != nil {
		log.Printf("[api] failed to register incident order_id=%s: %v", orderID, err)
		httpx.RespondError(w, http.StatusInternalServerError, "не удалось зарегистрировать инцидент")
		return
	}
	h.monitor.IncidentReported(*incident)
	log.Printf("[api] incident registered order_id=%s incident_id=%s", orderID, incident.ID)

	httpx.Respond(w, http.StatusCreated, models.IncidentResponse{
		IncidentID:   incident.ID,
		OrderID:      incident.OrderID,
		OperatorID:   incident.OperatorID,
		Status:       incident.Status,
		OrderStatus:  string(domain.StatusDispute),
		Reason:       incident.Reason,
		DamageAmount: incident.DamageAmount,
		Message:      "incident registered; payout calculation is handled by insurer system",
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
