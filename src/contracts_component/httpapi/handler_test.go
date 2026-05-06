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
