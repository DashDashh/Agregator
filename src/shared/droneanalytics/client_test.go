package droneanalytics

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestLogEventSendsEventLogPayload(t *testing.T) {
	var gotURL string
	var gotKey string
	var got []eventLogItem
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotURL = r.URL.String()
		gotKey = r.Header.Get("X-API-Key")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("Decode returned error: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"accepted":1}`))),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient(Config{
		Enabled:    true,
		BaseURL:    "https://infopanel.csse.ru/api/",
		APIKey:     "test-key",
		ServiceID:  7,
		APIVersion: "1.1.0",
		HTTPClient: &http.Client{Transport: transport},
	})

	err := client.LogEvent(context.Background(), Event{
		EventType: "safety_event",
		Severity:  "critical",
		Message:   "Security alert created",
		Timestamp: time.UnixMilli(123456789),
	})
	if err != nil {
		t.Fatalf("LogEvent returned error: %v", err)
	}
	if gotURL != "https://infopanel.csse.ru/api/log/event" {
		t.Fatalf("url = %q, want https://infopanel.csse.ru/api/log/event", gotURL)
	}
	if gotKey != "test-key" {
		t.Fatalf("X-API-Key = %q, want test-key", gotKey)
	}
	if len(got) != 1 {
		t.Fatalf("payload len = %d, want 1", len(got))
	}
	if got[0].EventType != "safety_event" || got[0].Service != "aggregator" || got[0].ServiceID != 7 {
		t.Fatalf("payload = %+v", got[0])
	}
	if got[0].Timestamp != 123456789 {
		t.Fatalf("timestamp = %d, want 123456789", got[0].Timestamp)
	}
}

func TestDisabledClientDoesNotCallServer(t *testing.T) {
	called := false
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader(nil))}, nil
	})

	client := NewClient(Config{
		Enabled:    false,
		BaseURL:    "https://infopanel.csse.ru/api",
		APIKey:     "test-key",
		HTTPClient: &http.Client{Transport: transport},
	})
	if err := client.LogEvent(context.Background(), Event{Message: "ignored"}); err != nil {
		t.Fatalf("LogEvent returned error: %v", err)
	}
	if called {
		t.Fatal("disabled client called server")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
