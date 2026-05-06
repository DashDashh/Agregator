package contracts_component

import (
	"encoding/json"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

func TestHandleConcludeContract(t *testing.T) {
	payload, err := json.Marshal(models.ConcludeContractRequest{
		OrderID:    "order-1",
		OperatorID: "operator-1",
		Price:      1500,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgConcludeContract,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept conclude_contract")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}
}

func TestHandleConfirmExecution(t *testing.T) {
	payload, err := json.Marshal(models.ConfirmExecutionRequest{
		ContractID: "contract-1",
		CustomerID: "customer-1",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgConfirmExecution,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept confirm_execution")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}

	body, ok := resp.Payload.(models.ConfirmExecutionResponse)
	if !ok {
		t.Fatalf("payload type = %T, want ConfirmExecutionResponse", resp.Payload)
	}
	if body.ContractID != "contract-1" || body.Status != "completed" {
		t.Fatalf("unexpected payload: %+v", body)
	}
}

func TestHandleCreateDispute(t *testing.T) {
	payload, err := json.Marshal(models.CreateDisputeRequest{
		ContractID:  "contract-1",
		CustomerID:  "customer-1",
		Description: "failed delivery",
		ClaimAmount: 500,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgCreateDispute,
		Payload:       payload,
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept create_dispute")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}

	body, ok := resp.Payload.(models.CreateDisputeResponse)
	if !ok {
		t.Fatalf("payload type = %T, want CreateDisputeResponse", resp.Payload)
	}
	if body.ContractID != "contract-1" || body.InsurancePayout != 500 {
		t.Fatalf("unexpected payload: %+v", body)
	}
}

func TestHandleRejectsUnknownAction(t *testing.T) {
	_, ok := NewHandler().Handle(models.Request{Action: models.MessageType("unknown")})
	if ok {
		t.Fatal("Handle accepted an unknown action")
	}
}

func TestHandleCreateDisputeRejectsInvalidPayload(t *testing.T) {
	resp, ok := NewHandler().Handle(models.Request{
		Action:        models.MsgCreateDispute,
		Payload:       []byte(`{bad json`),
		CorrelationID: "corr-1",
	})
	if !ok {
		t.Fatal("Handle did not accept create_dispute")
	}
	if resp.Success {
		t.Fatalf("expected invalid payload to fail: %+v", resp)
	}
}
