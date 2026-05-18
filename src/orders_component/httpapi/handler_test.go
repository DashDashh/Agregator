package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type fakeOrderStore struct {
	customer *store.Customer
	order    *store.Order
}

func (f *fakeOrderStore) GetCustomer(id string) (*store.Customer, bool) {
	if f.customer != nil && f.customer.ID == id {
		return f.customer, true
	}
	return nil, false
}

func (f *fakeOrderStore) SaveOrder(o *store.Order) error {
	f.order = o
	return nil
}

func (f *fakeOrderStore) GetOrder(id string) (*store.Order, bool) {
	if f.order != nil && f.order.ID == id {
		return f.order, true
	}
	return nil, false
}

func (f *fakeOrderStore) ListOrders() []*store.Order {
	if f.order == nil {
		return nil
	}
	return []*store.Order{f.order}
}

func (f *fakeOrderStore) ListOrdersByCustomer(customerID string) []*store.Order {
	if f.order != nil && f.order.CustomerID == customerID {
		return []*store.Order{f.order}
	}
	return nil
}

func (f *fakeOrderStore) UpdateOrderStatus(id string, status store.OrderStatus) bool {
	if f.order != nil && f.order.ID == id {
		f.order.Status = status
		return true
	}
	return false
}

type fakeOrderPublisher struct {
	published int
}

func (f *fakeOrderPublisher) PublishOrder(context.Context, *store.Order) error {
	f.published++
	return nil
}

func TestCreateOrderRequiresCustomerRole(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeOrderStore{}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"description":"docs"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCreateOrderRequiresAuth(t *testing.T) {
	h := NewHandler(&fakeOrderStore{}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"description":"docs"}`))
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestCreateOrderPublishesAndMarksSearching(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	repo := &fakeOrderStore{customer: &store.Customer{ID: "customer-1"}}
	pub := &fakeOrderPublisher{}
	h := NewHandler(repo, pub, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"description":"docs","budget":1000}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if pub.published != 1 {
		t.Fatalf("published = %d, want 1", pub.published)
	}
	if repo.order == nil || repo.order.Status != store.StatusSearching {
		t.Fatalf("order = %+v, want searching", repo.order)
	}
}

func TestCreateOrderAllowsNoAuthWhenAuthDisabled(t *testing.T) {
	repo := &fakeOrderStore{customer: &store.Customer{ID: "customer-1"}}
	pub := &fakeOrderPublisher{}
	h := NewHandlerWithAuthRequired(repo, pub, "test-secret", false)
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{
		"customer_id": "customer-1",
		"description": "docs",
		"budget": 1000
	}`))
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if repo.order == nil || repo.order.CustomerID != "customer-1" {
		t.Fatalf("order = %+v, want customer_id from payload", repo.order)
	}
	if pub.published != 1 {
		t.Fatalf("published = %d, want 1", pub.published)
	}
}

func TestCreateOrderRejectsMissingDescription(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeOrderStore{customer: &store.Customer{ID: "customer-1"}}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"budget":1000}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateOrderReturnsNotFoundForMissingCustomer(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeOrderStore{}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"description":"docs"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetOrderForbidsAnotherCustomerOrder(t *testing.T) {
	token, err := auth.NewToken("customer-2", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeOrderStore{order: &store.Order{
		ID:         "order-1",
		CustomerID: "customer-1",
	}}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.GetOrder(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestGetOrderReturnsNotFound(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeOrderStore{}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/orders/order-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.GetOrder(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestListOrdersForCustomerUsesTokenIdentity(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeOrderStore{order: &store.Order{
		ID:         "order-1",
		CustomerID: "customer-1",
	}}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.ListOrders(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "order-1") {
		t.Fatalf("body does not contain order: %s", rec.Body.String())
	}
}

func TestListOrdersFiltersByCustomerQueryForOperator(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeOrderStore{
		customer: &store.Customer{ID: "customer-1"},
		order:    &store.Order{ID: "order-1", CustomerID: "customer-1"},
	}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/orders?customer_id=customer-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.ListOrders(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "order-1") {
		t.Fatalf("body does not contain filtered order: %s", rec.Body.String())
	}
}

func TestListOrdersReturnsNotFoundForUnknownCustomerQuery(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeOrderStore{}, &fakeOrderPublisher{}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/orders?customer_id=missing", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.ListOrders(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
