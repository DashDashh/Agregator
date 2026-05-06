package componentbus

import (
	"strings"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type stubHandler struct {
	resp models.Response
	ok   bool
}

func (h stubHandler) Handle(models.Request) (models.Response, bool) {
	return h.resp, h.ok
}

func TestHandleRequestUsesHandlerResponse(t *testing.T) {
	resp, key := handleRequest("orders", models.Request{
		Action:        models.MsgCreateOrder,
		CorrelationID: "corr-1",
	}, "fallback", stubHandler{
		resp: models.Response{
			Action:        models.ResponseAction,
			Sender:        models.DefaultSender,
			CorrelationID: "resp-corr",
			Success:       true,
		},
		ok: true,
	})

	if !resp.Success {
		t.Fatal("handleRequest returned unsuccessful response")
	}
	if key != "resp-corr" {
		t.Fatalf("response key = %q, want %q", key, "resp-corr")
	}
}

func TestHandleRequestBuildsErrorForUnsupportedAction(t *testing.T) {
	resp, key := handleRequest("orders", models.Request{
		Action:        models.MsgGetAnalytics,
		CorrelationID: "corr-1",
		Timestamp:     "2026-05-06T00:00:00Z",
	}, "fallback", stubHandler{ok: false})

	if resp.Success {
		t.Fatal("handleRequest returned successful response for unsupported action")
	}
	if resp.Action != models.ResponseAction {
		t.Fatalf("response action = %q, want %q", resp.Action, models.ResponseAction)
	}
	if resp.CorrelationID != "corr-1" {
		t.Fatalf("correlation_id = %q, want %q", resp.CorrelationID, "corr-1")
	}
	if key != "corr-1" {
		t.Fatalf("response key = %q, want %q", key, "corr-1")
	}
	if !strings.Contains(resp.Error, "orders cannot handle action=get_analytics") {
		t.Fatalf("unexpected error: %q", resp.Error)
	}
}

func TestHandleRequestFallsBackToKafkaKey(t *testing.T) {
	resp, key := handleRequest("orders", models.Request{
		Action: models.MsgCreateOrder,
	}, "kafka-key", stubHandler{
		resp: models.Response{
			Action:  models.ResponseAction,
			Sender:  models.DefaultSender,
			Success: true,
		},
		ok: true,
	})

	if !resp.Success {
		t.Fatal("handleRequest returned unsuccessful response")
	}
	if key != "kafka-key" {
		t.Fatalf("response key = %q, want %q", key, "kafka-key")
	}
}
