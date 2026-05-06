package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type fakeContractStore struct {
	order *store.Order
}

func (f *fakeContractStore) GetOrder(id string) (*store.Order, bool) {
	if f.order != nil && f.order.ID == id {
		return f.order, true
	}
	return nil, false
}

func (f *fakeContractStore) ConfirmPrice(id, operatorID string, acceptedPrice, commissionAmount float64) bool {
	if f.order == nil || f.order.ID != id || f.order.OperatorID != operatorID {
		return false
	}
	f.order.Status = store.StatusConfirmed
	f.order.OfferedPrice = acceptedPrice
	f.order.CommissionAmount = commissionAmount
	return true
}

func (f *fakeContractStore) ConfirmCompletion(id string) bool {
	if f.order != nil && f.order.ID == id {
		f.order.Status = store.StatusCompleted
		return true
	}
	return false
}

func (f *fakeContractStore) SetOperatorOffer(orderID, operatorID string, price float64) bool {
	if f.order != nil && f.order.ID == orderID {
		f.order.OperatorID = operatorID
		f.order.OfferedPrice = price
		f.order.Status = store.StatusMatched
		return true
	}
	return false
}

type fakeContractPublisher struct {
	payload models.ConfirmPricePayload
	calls   int
}

func (f *fakeContractPublisher) PublishConfirmPrice(_ context.Context, payload models.ConfirmPricePayload) error {
	f.payload = payload
	f.calls++
	return nil
}

func TestConfirmPricePublishesPayload(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	repo := &fakeContractStore{order: &store.Order{
		ID:         "order-1",
		CustomerID: "customer-1",
		OperatorID: "operator-1",
		Status:     store.StatusMatched,
	}}
	pub := &fakeContractPublisher{}
	h := NewHandler(repo, pub, 0.1, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/confirm-price", strings.NewReader(`{
		"operator_id": "operator-1",
		"accepted_price": 1000
	}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.ConfirmPrice(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if pub.calls != 1 {
		t.Fatalf("publisher calls = %d, want 1", pub.calls)
	}
	if pub.payload.CommissionAmount != 100 {
		t.Fatalf("commission = %v, want 100", pub.payload.CommissionAmount)
	}
}

func TestConfirmPriceRejectsInvalidPayload(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeContractStore{order: &store.Order{
		ID:         "order-1",
		CustomerID: "customer-1",
		OperatorID: "operator-1",
	}}, &fakeContractPublisher{}, 0.1, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/confirm-price", strings.NewReader(`{
		"operator_id": "operator-1",
		"accepted_price": 0
	}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.ConfirmPrice(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestOfferPriceRequiresOperatorRole(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeContractStore{order: &store.Order{ID: "order-1"}}, &fakeContractPublisher{}, 0.1, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/offer", strings.NewReader(`{"price":1000}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.OfferPrice(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestOfferPriceStoresOperatorOffer(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	repo := &fakeContractStore{order: &store.Order{
		ID:         "order-1",
		CustomerID: "customer-1",
		Status:     store.StatusSearching,
	}}
	h := NewHandler(repo, &fakeContractPublisher{}, 0.1, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/offer", strings.NewReader(`{"price":1000}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.OfferPrice(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if repo.order.OperatorID != "operator-1" || repo.order.OfferedPrice != 1000 || repo.order.Status != store.StatusMatched {
		t.Fatalf("offer was not stored correctly: %+v", repo.order)
	}
}

func TestOfferPriceReturnsNotFoundForUnknownOrder(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeContractStore{}, &fakeContractPublisher{}, 0.1, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/offer", strings.NewReader(`{"price":1000}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.OfferPrice(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestConfirmCompletionRequiresOwner(t *testing.T) {
	token, err := auth.NewToken("customer-2", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeContractStore{order: &store.Order{
		ID:         "order-1",
		CustomerID: "customer-1",
		Status:     store.StatusCompletedPending,
	}}, &fakeContractPublisher{}, 0.1, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/confirm-completion", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.ConfirmCompletion(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestConfirmCompletionMarksOrderCompleted(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	repo := &fakeContractStore{order: &store.Order{
		ID:         "order-1",
		CustomerID: "customer-1",
		Status:     store.StatusCompletedPending,
	}}
	h := NewHandler(repo, &fakeContractPublisher{}, 0.1, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/orders/order-1/confirm-completion", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.ConfirmCompletion(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if repo.order.Status != store.StatusCompleted {
		t.Fatalf("status = %q, want %q", repo.order.Status, store.StatusCompleted)
	}
}
