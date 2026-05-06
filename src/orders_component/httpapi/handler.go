package httpapi

import (
	"context"
	"net/http"

	"github.com/kirilltahmazidi/aggregator/src/orders_component"
	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type Publisher interface {
	PublishOrder(ctx context.Context, order *store.Order) error
}

type Handler struct {
	store      orders_component.Store
	publisher  Publisher
	authSecret string
}

func NewHandler(s orders_component.Store, p Publisher, authSecret string) *Handler {
	return &Handler{store: s, publisher: p, authSecret: authSecret}
}

func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) (*auth.User, bool) {
	user, ok := auth.UserFromRequest(r, h.authSecret)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "нужна авторизация")
		return nil, false
	}
	return user, true
}
