package httpapi

import (
	"context"
	"net/http"

	"github.com/kirilltahmazidi/aggregator/src/orders_component"
	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/droneanalytics"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type Publisher interface {
	PublishOrder(ctx context.Context, order *store.Order) error
}

type Handler struct {
	store        orders_component.Store
	publisher    Publisher
	authSecret   string
	authRequired bool
	analytics    *droneanalytics.Client
}

func NewHandler(s orders_component.Store, p Publisher, authSecret string) *Handler {
	return NewHandlerWithAuthRequired(s, p, authSecret, true)
}

func NewHandlerWithAuthRequired(s orders_component.Store, p Publisher, authSecret string, authRequired bool) *Handler {
	return NewHandlerWithAuthRequiredAndAnalytics(s, p, authSecret, authRequired, nil)
}

func NewHandlerWithAuthRequiredAndAnalytics(s orders_component.Store, p Publisher, authSecret string, authRequired bool, analytics *droneanalytics.Client) *Handler {
	return &Handler{store: s, publisher: p, authSecret: authSecret, authRequired: authRequired, analytics: analytics}
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
