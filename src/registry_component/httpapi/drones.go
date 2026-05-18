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
)

func (h *Handler) RegisterDrone(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	operatorID, ok := operatorIDFromDronePath(r.URL.Path)
	if !ok {
		httpx.RespondError(w, http.StatusBadRequest, "id эксплуатанта не указан")
		return
	}
	if h.authRequired && (user.Role != "operator" || user.ID != operatorID) {
		httpx.RespondError(w, http.StatusForbidden, "нельзя добавлять дроны другому эксплуатанту")
		return
	}
	if _, found := h.store.GetOperator(operatorID); !found {
		httpx.RespondError(w, http.StatusNotFound, "эксплуатант не найден")
		return
	}

	var req struct {
		Name          string   `json:"name"`
		SecurityGoals []string `json:"security_goals"`
		Status        string   `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		httpx.RespondError(w, http.StatusBadRequest, "name обязателен")
		return
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "available"
	}

	drone := &domain.Drone{
		ID:            uuid.NewString(),
		OperatorID:    operatorID,
		Name:          strings.TrimSpace(req.Name),
		SecurityGoals: req.SecurityGoals,
		Status:        status,
		CreatedAt:     time.Now().UTC(),
	}
	if err := h.store.SaveDrone(drone); err != nil {
		log.Printf("[api] failed to save drone: %v", err)
		httpx.RespondError(w, http.StatusInternalServerError, "ошибка сохранения дрона")
		return
	}

	httpx.Respond(w, http.StatusCreated, drone)
}

func (h *Handler) ListDrones(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	operatorID, ok := operatorIDFromDronePath(r.URL.Path)
	if !ok {
		httpx.RespondError(w, http.StatusBadRequest, "id эксплуатанта не указан")
		return
	}
	if h.authRequired && (user.Role != "operator" || user.ID != operatorID) {
		httpx.RespondError(w, http.StatusForbidden, "нельзя смотреть дроны другого эксплуатанта")
		return
	}
	if _, found := h.store.GetOperator(operatorID); !found {
		httpx.RespondError(w, http.StatusNotFound, "эксплуатант не найден")
		return
	}

	httpx.Respond(w, http.StatusOK, h.store.ListDronesByOperator(operatorID))
}

func operatorIDFromDronePath(path string) (string, bool) {
	path = strings.TrimPrefix(path, "/operators/")
	path = strings.TrimSuffix(path, "/drones")
	path = strings.Trim(path, "/")
	return path, path != ""
}
