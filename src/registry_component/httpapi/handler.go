package httpapi

import (
	"net/http"

	"github.com/kirilltahmazidi/aggregator/src/registry_component"
	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
)

type Handler struct {
	store      registry_component.Store
	authSecret string
}

func NewHandler(s registry_component.Store, authSecret string) *Handler {
	return &Handler{store: s, authSecret: authSecret}
}

func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) (*auth.User, bool) {
	user, ok := auth.UserFromRequest(r, h.authSecret)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "нужна авторизация")
		return nil, false
	}
	return user, true
}
