package droneanalytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultAPIVersion = "1.1.0"
	defaultTimeout    = 3 * time.Second
)

type Config struct {
	Enabled    bool
	BaseURL    string
	APIKey     string
	ServiceID  int
	APIVersion string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type Client struct {
	enabled    bool
	baseURL    string
	apiKey     string
	serviceID  int
	apiVersion string
	httpClient *http.Client
}

type Event struct {
	EventType string
	Severity  string
	Message   string
	Timestamp time.Time
}

type eventLogItem struct {
	APIVersion string `json:"apiVersion"`
	Timestamp  int64  `json:"timestamp"`
	EventType  string `json:"event_type,omitempty"`
	Service    string `json:"service"`
	ServiceID  int    `json:"service_id"`
	Severity   string `json:"severity,omitempty"`
	Message    string `json:"message"`
}

func NewClient(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	apiVersion := strings.TrimSpace(cfg.APIVersion)
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}
	serviceID := cfg.ServiceID
	if serviceID <= 0 {
		serviceID = 1
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	return &Client{
		enabled:    cfg.Enabled && strings.TrimSpace(cfg.BaseURL) != "" && strings.TrimSpace(cfg.APIKey) != "",
		baseURL:    strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		apiKey:     strings.TrimSpace(cfg.APIKey),
		serviceID:  serviceID,
		apiVersion: apiVersion,
		httpClient: httpClient,
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled
}

func (c *Client) LogEventAsync(event Event) {
	if !c.Enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		if err := c.LogEvent(ctx, event); err != nil {
			log.Printf("[drone-analytics] failed to send event: %v", err)
		}
	}()
}

func (c *Client) LogEvent(ctx context.Context, event Event) error {
	if !c.Enabled() {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	timestamp := event.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	item := eventLogItem{
		APIVersion: c.apiVersion,
		Timestamp:  timestamp.UnixMilli(),
		EventType:  normalizeEventType(event.EventType),
		Service:    "aggregator",
		ServiceID:  c.serviceID,
		Severity:   normalizeSeverity(event.Severity),
		Message:    sanitizeMessage(event.Message),
	}

	body, err := json.Marshal([]eventLogItem{item})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/log/event", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}
	return nil
}

func normalizeEventType(v string) string {
	switch strings.TrimSpace(v) {
	case "safety_event":
		return "safety_event"
	default:
		return "event"
	}
}

func normalizeSeverity(v string) string {
	switch strings.TrimSpace(v) {
	case "debug", "info", "notice", "warning", "error", "critical", "alert", "emergency":
		return v
	default:
		return "info"
	}
}

func sanitizeMessage(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "Aggregator event"
	}
	if len(v) > 1024 {
		return v[:1024]
	}
	return v
}
