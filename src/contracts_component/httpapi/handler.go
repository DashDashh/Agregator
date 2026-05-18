package httpapi

import (
	"context"
	"net/http"

	"github.com/kirilltahmazidi/aggregator/src/contracts_component"
	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	securitymonitor "github.com/kirilltahmazidi/aggregator/src/security_monitor_component"
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
	authRequired   bool
	monitor        *securitymonitor.Monitor
}

func NewHandler(s contracts_component.Store, p Publisher, commissionRate float64, authSecret string) *Handler {
	return NewHandlerWithAuthRequired(s, p, commissionRate, authSecret, true)
}

func NewHandlerWithAuthRequired(s contracts_component.Store, p Publisher, commissionRate float64, authSecret string, authRequired bool) *Handler {
	return &Handler{
		store:          s,
		publisher:      p,
		commissionRate: commissionRate,
		authSecret:     authSecret,
		authRequired:   authRequired,
		monitor:        securitymonitor.New(securitySink(s)),
	}
}

func securitySink(s contracts_component.Store) securitymonitor.Sink {
	if alertStore, ok := s.(securitymonitor.AlertStore); ok {
		return securitymonitor.StoreSink{Store: alertStore, Next: securitymonitor.LogSink{}}
	}
	return securitymonitor.LogSink{}
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
