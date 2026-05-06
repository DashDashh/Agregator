package httpapi

import (
	"context"
	"net/http"

	"github.com/kirilltahmazidi/aggregator/src/contracts_component"
	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type Publisher interface {
	PublishConfirmPrice(ctx context.Context, payload models.ConfirmPricePayload) error
}

type Handler struct {
	store          contracts_component.Store
	publisher      Publisher
	commissionRate float64
	authSecret     string
}

func NewHandler(s contracts_component.Store, p Publisher, commissionRate float64, authSecret string) *Handler {
	return &Handler{store: s, publisher: p, commissionRate: commissionRate, authSecret: authSecret}
}

func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) (*auth.User, bool) {
	user, ok := auth.UserFromRequest(r, h.authSecret)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "нужна авторизация")
		return nil, false
	}
	return user, true
}
