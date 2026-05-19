package security_monitor_component

import "github.com/kirilltahmazidi/aggregator/src/shared/droneanalytics"

type AnalyticsSink struct {
	Client *droneanalytics.Client
	Next   Sink
}

func (s AnalyticsSink) Emit(alert Alert) {
	if s.Next != nil {
		s.Next.Emit(alert)
	}
	if s.Client == nil || !s.Client.Enabled() {
		return
	}
	s.Client.LogEventAsync(droneanalytics.Event{
		EventType: "safety_event",
		Severity:  mapAlertSeverity(alert.Severity),
		Message:   "Security alert: " + alert.Code + " for order " + alert.OrderID,
		Timestamp: alert.CreatedAt,
	})
}

func mapAlertSeverity(severity string) string {
	switch severity {
	case "high":
		return "critical"
	case "medium":
		return "warning"
	case "low":
		return "info"
	default:
		return "warning"
	}
}
