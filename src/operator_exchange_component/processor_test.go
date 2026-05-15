package operator_exchange_component

import (
	"encoding/json"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type fakeStore struct {
	order           *domain.Order
	incident        *domain.Incident
	offerOrderID    string
	offerOperatorID string
	offerPrice      float64
	resultOrderID   string
	resultSuccess   bool
	statusOrderID   string
	status          domain.OrderStatus
	apply           bool
}

func (f *fakeStore) GetOrder(id string) (*domain.Order, bool) {
	if f.order != nil && f.order.ID == id {
		return f.order, true
	}
	return nil, false
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

func (f *fakeStore) RegisterIncident(i *domain.Incident) error {
	f.incident = i
	return nil
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

func TestProcessOperatorMessageIgnoresPriceOfferWhenStoreRejects(t *testing.T) {
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

	result, err := ProcessOperatorMessage(&fakeStore{apply: false}, msg)
	if err != nil {
		t.Fatalf("ProcessOperatorMessage returned error: %v", err)
	}
	if result != ProcessIgnored {
		t.Fatalf("result = %q, want %q", result, ProcessIgnored)
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

func TestProcessOperatorMessageProcessesFailedOrderResult(t *testing.T) {
	payload, err := json.Marshal(models.OrderResultPayload{
		OrderID: "order-1",
		Success: false,
		Reason:  "drone_lost",
	})
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
	if store.resultOrderID != "order-1" || store.resultSuccess {
		t.Fatalf("stored result = %s/%v", store.resultOrderID, store.resultSuccess)
	}
}

func TestProcessOperatorMessageRegistersIncident(t *testing.T) {
	payload, err := json.Marshal(models.IncidentReportedPayload{
		OrderID:      "order-1",
		OperatorID:   "operator-1",
		Reason:       "drone_lost",
		DamageAmount: 1500,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	msg, err := json.Marshal(models.Request{Action: models.MsgIncidentReported, Payload: payload})
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
	if store.incident == nil {
		t.Fatal("incident was not registered")
	}
	if store.incident.OrderID != "order-1" || store.incident.Reason != "drone_lost" || store.incident.DamageAmount != 1500 {
		t.Fatalf("incident = %+v", store.incident)
	}
}

func TestProcessOperatorMessageRejectsInvalidPayload(t *testing.T) {
	msg := []byte(`{"action":"price_offer","payload":"oops"}`)

	result, err := ProcessOperatorMessage(&fakeStore{}, msg)
	if err == nil {
		t.Fatal("expected invalid payload error")
	}
	if result != ProcessInvalidPayload {
		t.Fatalf("result = %q, want %q", result, ProcessInvalidPayload)
	}
}

func TestProcessOperatorMessageRejectsInvalidIncidentPayload(t *testing.T) {
	payload, err := json.Marshal(models.IncidentReportedPayload{OrderID: "order-1"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	msg, err := json.Marshal(models.Request{Action: models.MsgIncidentReported, Payload: payload})
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}

	result, err := ProcessOperatorMessage(&fakeStore{}, msg)
	if err == nil {
		t.Fatal("expected invalid incident payload error")
	}
	if result != ProcessInvalidPayload {
		t.Fatalf("result = %q, want %q", result, ProcessInvalidPayload)
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
