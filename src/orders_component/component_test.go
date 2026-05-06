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

func TestHandleRejectsUnknownAction(t *testing.T) {
	_, ok := NewHandler().Handle(models.Request{Action: models.MessageType("unknown")})
	if ok {
		t.Fatal("Handle accepted an unknown action")
	}
}
