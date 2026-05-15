package models

import (
	"encoding/json"
	"testing"
)

func TestRequestUnmarshalSupportsTypeAndRequestIDAliases(t *testing.T) {
	var req Request
	err := json.Unmarshal([]byte(`{
		"type": "create_order",
		"payload": {"description": "docs"},
		"request_id": "req-1"
	}`), &req)
	if err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}

	if req.Action != MsgCreateOrder {
		t.Fatalf("Action = %q, want %q", req.Action, MsgCreateOrder)
	}
	if req.CorrelationID != "req-1" {
		t.Fatalf("CorrelationID = %q, want req-1", req.CorrelationID)
	}
	if req.Timestamp == "" {
		t.Fatal("Timestamp was not normalized")
	}
}

func TestRequestUnmarshalPrefersActionAndCorrelationID(t *testing.T) {
	var req Request
	err := json.Unmarshal([]byte(`{
		"action": "register_customer",
		"type": "create_order",
		"correlation_id": "corr-1",
		"request_id": "req-1"
	}`), &req)
	if err != nil {
		t.Fatalf("UnmarshalJSON returned error: %v", err)
	}

	if req.Action != MsgRegisterCustomer {
		t.Fatalf("Action = %q, want %q", req.Action, MsgRegisterCustomer)
	}
	if req.CorrelationID != "corr-1" {
		t.Fatalf("CorrelationID = %q, want corr-1", req.CorrelationID)
	}
}
