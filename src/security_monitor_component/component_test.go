package security_monitor_component

import (
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type captureSink struct {
	alerts []Alert
}

func (s *captureSink) Emit(alert Alert) {
	s.alerts = append(s.alerts, alert)
}

func TestIncidentReportedEmitsHighSeverityAlert(t *testing.T) {
	sink := &captureSink{}
	monitor := New(sink)

	alert := monitor.IncidentReported(domain.Incident{OrderID: "order-1"})

	if alert.Code != "incident_reported" || alert.Severity != "high" || alert.OrderID != "order-1" {
		t.Fatalf("alert = %+v", alert)
	}
	if len(sink.alerts) != 1 {
		t.Fatalf("sink alerts = %d, want 1", len(sink.alerts))
	}
}

func TestOrderFailedEmitsAlertWithReason(t *testing.T) {
	sink := &captureSink{}
	monitor := New(sink)

	alert := monitor.OrderFailed(models.OrderResultPayload{
		OrderID: "order-1",
		Reason:  "drone_lost",
	})

	if alert.Code != "operator_reported_failure" || alert.Severity != "high" || alert.Message != "drone_lost" {
		t.Fatalf("alert = %+v", alert)
	}
	if len(sink.alerts) != 1 {
		t.Fatalf("sink alerts = %d, want 1", len(sink.alerts))
	}
}

func TestOrderFailedUsesDefaultMessage(t *testing.T) {
	sink := &captureSink{}
	monitor := New(sink)

	alert := monitor.OrderFailed(models.OrderResultPayload{OrderID: "order-1"})

	if alert.Message == "" {
		t.Fatal("default failure message is empty")
	}
}

func TestPriceOfferReceivedAlertsOnMissingSecurityGoals(t *testing.T) {
	sink := &captureSink{}
	monitor := New(sink)

	alerts := monitor.PriceOfferReceived(
		&domain.Order{ID: "order-1", SecurityGoals: []string{"CB1", "CB2"}},
		models.PriceOfferPayload{OrderID: "order-1", ProvidedSecurityGoals: []string{"CB1"}},
	)

	if len(alerts) != 1 {
		t.Fatalf("alerts = %d, want 1", len(alerts))
	}
	if alerts[0].Code != "security_goals_not_covered" {
		t.Fatalf("alert code = %q", alerts[0].Code)
	}
}

func TestPriceOfferReceivedIgnoresCoveredSecurityGoals(t *testing.T) {
	sink := &captureSink{}
	monitor := New(sink)

	alerts := monitor.PriceOfferReceived(
		&domain.Order{ID: "order-1", SecurityGoals: []string{"CB1"}},
		models.PriceOfferPayload{OrderID: "order-1", ProvidedSecurityGoals: []string{"CB1"}},
	)

	if len(alerts) != 0 {
		t.Fatalf("alerts = %d, want 0", len(alerts))
	}
	if len(sink.alerts) != 0 {
		t.Fatalf("sink alerts = %d, want 0", len(sink.alerts))
	}
}
