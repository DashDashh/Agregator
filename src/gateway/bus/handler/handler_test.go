package handler

import (
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

func TestHandleRoutesKnownAction(t *testing.T) {
	resp := New().Handle(models.Request{
		Action:        models.MsgGetAnalytics,
		Payload:       []byte(`{}`),
		CorrelationID: "corr-1",
	})

	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}
	if resp.CorrelationID != "corr-1" {
		t.Fatalf("CorrelationID = %q, want corr-1", resp.CorrelationID)
	}
}

func TestHandleRejectsUnknownAction(t *testing.T) {
	resp := New().Handle(models.Request{
		Action:        models.MessageType("unknown"),
		CorrelationID: "corr-1",
	})

	if resp.Success {
		t.Fatalf("expected failure response: %+v", resp)
	}
	if resp.Error == "" {
		t.Fatal("expected error message")
	}
}
