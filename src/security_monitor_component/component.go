package security_monitor_component

import (
	"log"

	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type Alert struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	OrderID  string `json:"order_id,omitempty"`
	Message  string `json:"message"`
}

type Sink interface {
	Emit(alert Alert)
}

type LogSink struct{}

func (LogSink) Emit(alert Alert) {
	log.Printf("[security-monitor] severity=%s code=%s order_id=%s message=%s",
		alert.Severity, alert.Code, alert.OrderID, alert.Message)
}

type Monitor struct {
	sink Sink
}

func New(sink Sink) *Monitor {
	if sink == nil {
		sink = LogSink{}
	}
	return &Monitor{sink: sink}
}

func (m *Monitor) IncidentReported(i domain.Incident) Alert {
	alert := Alert{
		Code:     "incident_reported",
		Severity: "high",
		OrderID:  i.OrderID,
		Message:  "negative order scenario registered",
	}
	m.sink.Emit(alert)
	return alert
}

func (m *Monitor) OrderFailed(payload models.OrderResultPayload) Alert {
	alert := Alert{
		Code:     "operator_reported_failure",
		Severity: "high",
		OrderID:  payload.OrderID,
		Message:  payload.Reason,
	}
	if alert.Message == "" {
		alert.Message = "operator reported failed order result"
	}
	m.sink.Emit(alert)
	return alert
}

func (m *Monitor) PriceOfferReceived(order *domain.Order, payload models.PriceOfferPayload) []Alert {
	if order == nil || len(order.SecurityGoals) == 0 {
		return nil
	}

	provided := make(map[string]struct{}, len(payload.ProvidedSecurityGoals))
	for _, goal := range payload.ProvidedSecurityGoals {
		provided[goal] = struct{}{}
	}

	var missing []string
	for _, goal := range order.SecurityGoals {
		if _, ok := provided[goal]; !ok {
			missing = append(missing, goal)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	alert := Alert{
		Code:     "security_goals_not_covered",
		Severity: "medium",
		OrderID:  order.ID,
		Message:  "operator offer does not cover all requested security goals",
	}
	m.sink.Emit(alert)
	return []Alert{alert}
}
