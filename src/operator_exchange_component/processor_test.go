package operator_exchange_component

import (
	"encoding/json"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type fakeStore struct {
	offerOrderID    string
	offerOperatorID string
	offerPrice      float64
	resultOrderID   string
	resultSuccess   bool
	statusOrderID   string
	status          domain.OrderStatus
	apply           bool
}

func (f *fakeStore) SetOperatorOffer(orderID, operatorID string, price float64) bool {
	f.offerOrderID = orderID
	f.offerOperatorID = operatorID
	f.offerPrice = price
	return f.apply
}

func (f *fakeStore) ProcessOrderResult(orderID string, success bool) bool {
	f.resultOrderID = orderID
	f.resultSuccess = success
	return f.apply
}

func (f *fakeStore) UpdateOrderStatus(id string, status domain.OrderStatus) bool {
	f.statusOrderID = id
	f.status = status
	return f.apply
}

func TestProcessOperatorMessageAppliesPriceOffer(t *testing.T) {
	payload, err := json.Marshal(models.PriceOfferPayload{
		OrderID:    "order-1",
		OperatorID: "operator-1",
		Price:      1200,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	msg, err := json.Marshal(models.Request{Action: models.MsgPriceOffer, Payload: payload})
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	store := &fakeStore{apply: true}

	result, err := ProcessOperatorMessage(store, msg)
	if err != nil {
		t.Fatalf("ProcessOperatorMessage returned error: %v", err)
	}
	if result != ProcessApplied {
		t.Fatalf("result = %q, want %q", result, ProcessApplied)
	}
	if store.offerOrderID != "order-1" || store.offerOperatorID != "operator-1" || store.offerPrice != 1200 {
		t.Fatalf("stored offer = %s/%s/%v", store.offerOrderID, store.offerOperatorID, store.offerPrice)
	}
}

func TestProcessOperatorMessageAppliesOrderResult(t *testing.T) {
	payload, err := json.Marshal(models.OrderResultPayload{OrderID: "order-1", Success: true})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	msg, err := json.Marshal(models.Request{Action: models.MsgOrderResult, Payload: payload})
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	store := &fakeStore{apply: true}

	result, err := ProcessOperatorMessage(store, msg)
	if err != nil {
		t.Fatalf("ProcessOperatorMessage returned error: %v", err)
	}
	if result != ProcessApplied {
		t.Fatalf("result = %q, want %q", result, ProcessApplied)
	}
	if store.resultOrderID != "order-1" || !store.resultSuccess {
		t.Fatalf("stored result = %s/%v", store.resultOrderID, store.resultSuccess)
	}
}

func TestProcessOperatorMessageReportsInvalidMessage(t *testing.T) {
	result, err := ProcessOperatorMessage(&fakeStore{}, []byte(`{bad json`))
	if err == nil {
		t.Fatal("expected invalid message error")
	}
	if result != ProcessInvalidMessage {
		t.Fatalf("result = %q, want %q", result, ProcessInvalidMessage)
	}
}

func TestProcessOperatorMessageReportsUnknownAction(t *testing.T) {
	msg, err := json.Marshal(models.Request{Action: models.MessageType("unknown")})
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}

	result, err := ProcessOperatorMessage(&fakeStore{}, msg)
	if err != nil {
		t.Fatalf("ProcessOperatorMessage returned error: %v", err)
	}
	if result != ProcessUnknownAction {
		t.Fatalf("result = %q, want %q", result, ProcessUnknownAction)
	}
}
