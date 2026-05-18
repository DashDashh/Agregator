package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
)

type fakeStore struct {
	alerts     []*domain.SecurityAlert
	resolvedID string
}

func (f *fakeStore) ListSecurityAlerts(string, int) []*domain.SecurityAlert {
	return f.alerts
}

func (f *fakeStore) ResolveSecurityAlert(id string) bool {
	f.resolvedID = id
	return id == "alert-1"
}

func TestListAlertsRequiresOperator(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/security/alerts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	NewHandler(&fakeStore{}, "test-secret").ListAlerts(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestListAlertsReturnsStoredAlerts(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	repo := &fakeStore{alerts: []*domain.SecurityAlert{{
		ID:        "alert-1",
		Code:      "incident_reported",
		Severity:  "high",
		Source:    "incident",
		Message:   "negative order scenario registered",
		Status:    "open",
		CreatedAt: time.Now().UTC(),
	}}}

	req := httptest.NewRequest(http.MethodGet, "/security/alerts?status=open", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	NewHandler(repo, "test-secret").ListAlerts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestResolveAlert(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	repo := &fakeStore{}

	req := httptest.NewRequest(http.MethodPost, "/security/alerts/alert-1/resolve", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	NewHandler(repo, "test-secret").ResolveAlert(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if repo.resolvedID != "alert-1" {
		t.Fatalf("resolvedID = %q, want alert-1", repo.resolvedID)
	}
}
