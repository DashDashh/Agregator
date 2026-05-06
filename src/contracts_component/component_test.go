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
