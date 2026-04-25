package handlers

import (
	"context"
	"net/http"

	"github.com/kirilltahmazidi/aggregator/internal/bus/models"
	"github.com/kirilltahmazidi/aggregator/internal/store"
)

// Publisher sends order workflow messages to external operator transports.
type Publisher interface {
	PublishOrder(ctx context.Context, order *store.Order) error
	PublishConfirmPrice(ctx context.Context, payload models.ConfirmPricePayload) error
}

// Handler contains REST API dependencies.
type Handler struct {
	store          *store.Store
	publisher      Publisher
	commissionRate float64
	authSecret     string
}

func NewHandler(s *store.Store, p Publisher, commissionRate float64, authSecret string) *Handler {
	return &Handler{store: s, publisher: p, commissionRate: commissionRate, authSecret: authSecret}
}

// Health checks that the service is alive.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}
