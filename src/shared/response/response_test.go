package response

import (
	"strings"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

func TestOkBuildsSuccessfulResponse(t *testing.T) {
	resp := Ok(models.Request{CorrelationID: "corr-1"}, map[string]string{"ok": "yes"})
	if !resp.Success {
		t.Fatal("Ok returned unsuccessful response")
	}
	if resp.Action != models.ResponseAction {
		t.Fatalf("Action = %q, want %q", resp.Action, models.ResponseAction)
	}
	if resp.CorrelationID != "corr-1" {
		t.Fatalf("CorrelationID = %q, want corr-1", resp.CorrelationID)
	}
}

func TestErrorfFormatsFailureResponse(t *testing.T) {
	resp := Errorf("component", models.Request{CorrelationID: "corr-1"}, "bad %s", "payload")
	if resp.Success {
		t.Fatal("Errorf returned successful response")
	}
	if !strings.Contains(resp.Error, "bad payload") {
		t.Fatalf("Error = %q, want formatted message", resp.Error)
	}
}
