package orders_component

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type fakeStore struct {
	customer *domain.Customer
	order    *domain.Order
	drone    *domain.Drone
	operator *domain.Operator
}

func (f *fakeStore) GetCustomer(id string) (*domain.Customer, bool) {
	if f.customer != nil && f.customer.ID == id {
		return f.customer, true
	}
	return nil, false
}

func (f *fakeStore) SaveOrder(o *domain.Order) error {
	f.order = o
	return nil
}

func (f *fakeStore) GetOrder(id string) (*domain.Order, bool) {
	if f.order != nil && f.order.ID == id {
		return f.order, true
	}
	return nil, false
}

func (f *fakeStore) ListOrders() []*domain.Order {
	if f.order == nil {
		return nil
	}
	return []*domain.Order{f.order}
}

func (f *fakeStore) ListOrdersByCustomer(customerID string) []*domain.Order {
	if f.order != nil && f.order.CustomerID == customerID {
		return []*domain.Order{f.order}
	}
	return nil
}

func (f *fakeStore) UpdateOrderStatus(id string, status domain.OrderStatus) bool {
	if f.order != nil && f.order.ID == id {
		f.order.Status = status
		return true
	}
	return false
}

func (f *fakeStore) FindExecutorDrone([]string) (*domain.Drone, *domain.Operator, bool) {
	if f.drone == nil || f.operator == nil {
		return nil, nil, false
	}
	return f.drone, f.operator, true
}

func TestHandleCreateOrder(t *testing.T) {
	payload, err := json.Marshal(models.CreateOrderRequest{
		CustomerID:  "customer-1",
		Description: "deliver docs",
		Budget:      1200,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgCreateOrder,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept create_order")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}
	if resp.CorrelationID != "corr-1" {
		t.Fatalf("CorrelationID = %q, want corr-1", resp.CorrelationID)
	}
}

func TestHandleSelectExecutor(t *testing.T) {
	payload, err := json.Marshal(models.SelectExecutorRequest{
		OrderID:    "order-1",
		OperatorID: "operator-1",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgSelectExecutor,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept select_executor")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}

	body, ok := resp.Payload.(models.SelectExecutorResponse)
	if !ok {
		t.Fatalf("payload type = %T, want SelectExecutorResponse", resp.Payload)
	}
	if body.OrderID != "order-1" || body.OperatorID != "operator-1" || body.Status != "executor_selected" {
		t.Fatalf("unexpected payload: %+v", body)
	}
}

func TestHandleAutoSearchExecutor(t *testing.T) {
	payload, err := json.Marshal(models.AutoSearchExecutorRequest{
		OrderID:   "order-1",
		MaxBudget: 1000,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgAutoSearchExecutor,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept auto_search_executor")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}

	body, ok := resp.Payload.(models.AutoSearchExecutorResponse)
	if !ok {
		t.Fatalf("payload type = %T, want AutoSearchExecutorResponse", resp.Payload)
	}
	if body.OrderID != "order-1" || len(body.Candidates) != 2 {
		t.Fatalf("unexpected payload: %+v", body)
	}
}

func TestHandleAutoSearchExecutorSelectsMatchingDrone(t *testing.T) {
	payload, err := json.Marshal(models.AutoSearchExecutorRequest{OrderID: "order-1"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	store := &fakeStore{
		order: &domain.Order{ID: "order-1", SecurityGoals: []string{"CB1", "CB2"}},
		drone: &domain.Drone{
			ID:            "drone-1",
			OperatorID:    "operator-1",
			Name:          "Drone Alpha",
			SecurityGoals: []string{"CB1", "CB2", "CB3"},
			Status:        "available",
			CreatedAt:     time.Now().UTC(),
		},
		operator: &domain.Operator{ID: "operator-1", Name: "Operator Alpha"},
	}

	resp, ok := NewStoreHandler(store).Handle(models.Request{
		Action:        models.MsgAutoSearchExecutor,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept auto_search_executor")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}
	body, ok := resp.Payload.(models.AutoSearchExecutorResponse)
	if !ok {
		t.Fatalf("payload type = %T, want AutoSearchExecutorResponse", resp.Payload)
	}
	if body.Selected == nil || body.Selected.DroneID != "drone-1" || body.Selected.OperatorID != "operator-1" {
		t.Fatalf("unexpected selected candidate: %+v", body.Selected)
	}
}

func TestHandleAutoSearchExecutorFailsWhenNoDroneMatches(t *testing.T) {
	payload, err := json.Marshal(models.AutoSearchExecutorRequest{OrderID: "order-1"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewStoreHandler(&fakeStore{order: &domain.Order{ID: "order-1"}}).Handle(models.Request{
		Action:        models.MsgAutoSearchExecutor,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept auto_search_executor")
	}
	if resp.Success {
		t.Fatalf("expected no matching drone to fail: %+v", resp)
	}
}

func TestHandleCreateOrderRejectsInvalidPayload(t *testing.T) {
	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgCreateOrder,
		Payload:       []byte(`{bad json`),
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept create_order")
	}
	if resp.Success {
		t.Fatalf("expected invalid payload to fail: %+v", resp)
	}
}

func TestHandleRejectsUnknownAction(t *testing.T) {
	_, ok := NewHandler().Handle(models.Request{Action: models.MessageType("unknown")})
	if ok {
		t.Fatal("Handle accepted an unknown action")
	}
}

func TestHandles(t *testing.T) {
	if !Handles(models.MsgCreateOrder) || !Handles(models.MsgSelectExecutor) || !Handles(models.MsgAutoSearchExecutor) {
		t.Fatal("Handles rejected orders actions")
	}
	if Handles(models.MsgRegisterCustomer) {
		t.Fatal("Handles accepted non-orders action")
	}
}
