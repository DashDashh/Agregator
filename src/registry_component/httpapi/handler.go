package httpapi

import (
	"net/http"

	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type RegistryStore interface {
	SaveCustomer(c *store.Customer) error
	GetCustomer(id string) (*store.Customer, bool)
	GetCustomerByEmail(email string) (*store.Customer, bool)
	SetCustomerPasswordHash(id, passwordHash string) bool
	SaveOperator(op *store.Operator) error
	GetOperator(id string) (*store.Operator, bool)
	GetOperatorByEmail(email string) (*store.Operator, bool)
	SetOperatorPasswordHash(id, passwordHash string) bool
}

type Handler struct {
	store      RegistryStore
	authSecret string
}

func NewHandler(s RegistryStore, authSecret string) *Handler {
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
