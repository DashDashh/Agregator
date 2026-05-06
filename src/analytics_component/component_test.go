package analytics_component

import (
	"encoding/json"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

func TestHandleGetAnalytics(t *testing.T) {
	payload, err := json.Marshal(models.GetAnalyticsRequest{From: "2026-05-01T00:00:00Z", To: "2026-05-06T00:00:00Z"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	resp, ok := NewHandler().Handle(models.Request{
		Action:  models.MsgGetAnalytics,
		Payload: payload,
	})
	if !ok {
		t.Fatal("Handle did not accept get_analytics")
	}
	if !resp.Success {
		t.Fatalf("response failed: %+v", resp)
	}
}

func TestHandleGetAnalyticsRejectsUnknownAction(t *testing.T) {
	_, ok := NewHandler().Handle(models.Request{Action: models.MsgCreateOrder})
	if ok {
		t.Fatal("Handle accepted an unsupported action")
	}
}
