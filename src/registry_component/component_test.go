package registry_component

import (
	"encoding/json"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

func TestHandleRegisterCustomer(t *testing.T) {
	payload, err := json.Marshal(models.RegisterCustomerRequest{
		Name:  "Ivan",
		Email: "ivan@example.com",
		Phone: "+79001234567",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgRegisterCustomer,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept register_customer")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}
}

func TestHandleRegisterOperator(t *testing.T) {
	payload, err := json.Marshal(models.RegisterOperatorRequest{
		Name:    "Operator",
		License: "LIC-1",
		Email:   "op@example.com",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgRegisterOperator,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept register_operator")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}
}

func TestHandleRegisterOperatorRejectsInvalidPayload(t *testing.T) {
	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgRegisterOperator,
		Payload:       []byte(`{bad json`),
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept register_operator")
	}
	if resp.Success {
		t.Fatalf("expected invalid payload to fail: %+v", resp)
	}
}

func TestHandles(t *testing.T) {
	if !Handles(models.MsgRegisterCustomer) || !Handles(models.MsgRegisterOperator) {
		t.Fatal("Handles rejected registry actions")
	}
	if Handles(models.MsgCreateOrder) {
		t.Fatal("Handles accepted non-registry action")
	}
}
