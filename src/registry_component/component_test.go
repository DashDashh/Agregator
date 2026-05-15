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
