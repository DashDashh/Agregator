package httpapi

import (
	"context"
	"net/http"

	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type Publisher interface {
	PublishConfirmPrice(ctx context.Context, payload models.ConfirmPricePayload) error
}

type ContractStore interface {
	GetOrder(id string) (*store.Order, bool)
	ConfirmPrice(id, operatorID string, acceptedPrice, commissionAmount float64) bool
	ConfirmCompletion(id string) bool
	SetOperatorOffer(orderID, operatorID string, price float64) bool
}

type Handler struct {
	store          ContractStore
	publisher      Publisher
	commissionRate float64
	authSecret     string
}

func NewHandler(s ContractStore, p Publisher, commissionRate float64, authSecret string) *Handler {
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
