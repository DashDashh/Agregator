package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
)

type Store interface {
	ListSecurityAlerts(status string, limit int) []*domain.SecurityAlert
	ResolveSecurityAlert(id string) bool
}

type Handler struct {
	store      Store
	authSecret string
}

func NewHandler(store Store, authSecret string) *Handler {
	return &Handler{store: store, authSecret: authSecret}
}

func (h *Handler) requireOperator(w http.ResponseWriter, r *http.Request) bool {
	user, ok := auth.UserFromRequest(r, h.authSecret)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "нужна авторизация")
		return false
	}
	if user.Role != "operator" {
		httpx.RespondError(w, http.StatusForbidden, "security alerts доступны только эксплуатантам")
		return false
	}
	return true
}

func (h *Handler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	if !h.requireOperator(w, r) {
		return
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	limit := 100
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	httpx.Respond(w, http.StatusOK, h.store.ListSecurityAlerts(status, limit))
}

func (h *Handler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	if !h.requireOperator(w, r) {
		return
	}
	id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/security/alerts/"), "/resolve")
	if id == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id алерта не указан")
		return
	}
	if !h.store.ResolveSecurityAlert(id) {
		httpx.RespondError(w, http.StatusNotFound, "alert не найден или уже закрыт")
		return
	}
	httpx.Respond(w, http.StatusOK, map[string]string{"id": id, "status": "resolved"})
}
