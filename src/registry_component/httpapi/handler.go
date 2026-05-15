package httpapi

import (
	"net/http"

	"github.com/kirilltahmazidi/aggregator/src/registry_component"
	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
)

type Handler struct {
	store        registry_component.Store
	authSecret   string
	authRequired bool
}

func NewHandler(s registry_component.Store, authSecret string) *Handler {
	return NewHandlerWithAuthRequired(s, authSecret, true)
}

func NewHandlerWithAuthRequired(s registry_component.Store, authSecret string, authRequired bool) *Handler {
	return &Handler{store: s, authSecret: authSecret, authRequired: authRequired}
}

func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) (*auth.User, bool) {
	user, ok := auth.UserFromRequest(r, h.authSecret)
	if !ok {
		if !h.authRequired {
			return &auth.User{Role: "integration"}, true
		}
		httpx.RespondError(w, http.StatusUnauthorized, "нужна авторизация")
		return nil, false
	}
	return user, true
}
