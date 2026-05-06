package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRespondWritesJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	Respond(rec, http.StatusCreated, map[string]string{"status": "ok"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content-type = %q", rec.Header().Get("Content-Type"))
	}
	if strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRespondErrorWritesErrorPayload(t *testing.T) {
	rec := httptest.NewRecorder()

	RespondError(rec, http.StatusForbidden, "нет доступа")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if !strings.Contains(rec.Body.String(), "нет доступа") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
