package orders_component

import (
	"encoding/json"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

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
