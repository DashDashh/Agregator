package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	contractsapi "github.com/kirilltahmazidi/aggregator/src/contracts_component/httpapi"
	ordersapi "github.com/kirilltahmazidi/aggregator/src/orders_component/httpapi"
	registryapi "github.com/kirilltahmazidi/aggregator/src/registry_component/httpapi"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type routerStore struct {
	customer *store.Customer
	operator *store.Operator
	order    *store.Order
}

func (s *routerStore) SaveCustomer(c *store.Customer) error {
	s.customer = c
	return nil
}

func (s *routerStore) GetCustomer(id string) (*store.Customer, bool) {
	if s.customer != nil && s.customer.ID == id {
		return s.customer, true
	}
	return nil, false
}

func (s *routerStore) GetCustomerByEmail(email string) (*store.Customer, bool) {
	if s.customer != nil && s.customer.Email == email {
		return s.customer, true
	}
	return nil, false
}

func (s *routerStore) SetCustomerPasswordHash(id, passwordHash string) bool {
	if s.customer != nil && s.customer.ID == id {
		s.customer.PasswordHash = passwordHash
		return true
	}
	return false
}

func (s *routerStore) SaveOperator(op *store.Operator) error {
	s.operator = op
	return nil
}

func (s *routerStore) GetOperator(id string) (*store.Operator, bool) {
	if s.operator != nil && s.operator.ID == id {
		return s.operator, true
	}
	return nil, false
}

func (s *routerStore) GetOperatorByEmail(email string) (*store.Operator, bool) {
	if s.operator != nil && s.operator.Email == email {
		return s.operator, true
	}
	return nil, false
}

func (s *routerStore) SetOperatorPasswordHash(id, passwordHash string) bool {
	if s.operator != nil && s.operator.ID == id {
		s.operator.PasswordHash = passwordHash
		return true
	}
	return false
}

func (s *routerStore) SaveOrder(o *store.Order) error {
	s.order = o
	return nil
}

func (s *routerStore) GetOrder(id string) (*store.Order, bool) {
	if s.order != nil && s.order.ID == id {
		return s.order, true
	}
	return nil, false
}

func (s *routerStore) ListOrders() []*store.Order {
	if s.order == nil {
		return nil
	}
	return []*store.Order{s.order}
}

func (s *routerStore) ListOrdersByCustomer(customerID string) []*store.Order {
	if s.order != nil && s.order.CustomerID == customerID {
		return []*store.Order{s.order}
	}
	return nil
}

func (s *routerStore) UpdateOrderStatus(id string, status store.OrderStatus) bool {
	if s.order != nil && s.order.ID == id {
		s.order.Status = status
		return true
	}
	return false
}

func (s *routerStore) ConfirmPrice(id, operatorID string, acceptedPrice, commissionAmount float64) bool {
	if s.order != nil && s.order.ID == id {
		s.order.OperatorID = operatorID
		s.order.OfferedPrice = acceptedPrice
		s.order.CommissionAmount = commissionAmount
		return true
	}
	return false
}

func (s *routerStore) ConfirmCompletion(id string) bool {
	if s.order != nil && s.order.ID == id {
		s.order.Status = store.StatusCompleted
		return true
	}
	return false
}

func (s *routerStore) SetOperatorOffer(orderID, operatorID string, price float64) bool {
	if s.order != nil && s.order.ID == orderID {
		s.order.OperatorID = operatorID
		s.order.OfferedPrice = price
		s.order.Status = store.StatusMatched
		return true
	}
	return false
}

type routerPublisher struct{}

func (p routerPublisher) PublishOrder(context.Context, *store.Order) error {
	return nil
}

func (p routerPublisher) PublishConfirmPrice(context.Context, models.ConfirmPricePayload) error {
	return nil
}

func newTestRouter() http.Handler {
	repo := &routerStore{}
	pub := routerPublisher{}
	return NewRouter(Handlers{
		Registry:  registryapi.NewHandler(repo, "test-secret"),
		Orders:    ordersapi.NewHandler(repo, pub, "test-secret"),
		Contracts: contractsapi.NewHandler(repo, pub, 0.1, "test-secret"),
	})
}

func TestRouterHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("CORS header is missing")
	}
}

func TestRouterRejectsUnsupportedMethod(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)

	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestRouterDispatchesCustomerRegistration(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{
		"name": "Ivan",
		"email": "ivan-router@example.com",
		"password": "strongpass123",
		"phone": "+79001234567"
	}`))

	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("body does not contain token: %s", rec.Body.String())
	}
}
