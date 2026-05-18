package security_monitor_component

import (
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type Alert struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Severity  string    `json:"severity"`
	Source    string    `json:"source"`
	OrderID   string    `json:"order_id,omitempty"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Sink interface {
	Emit(alert Alert)
}

type LogSink struct{}

func (LogSink) Emit(alert Alert) {
	log.Printf("[security-monitor] id=%s severity=%s code=%s source=%s order_id=%s message=%s",
		alert.ID, alert.Severity, alert.Code, alert.Source, alert.OrderID, alert.Message)
}

type AlertStore interface {
	SaveSecurityAlert(alert *domain.SecurityAlert) error
}

type StoreSink struct {
	Store AlertStore
	Next  Sink
}

func (s StoreSink) Emit(alert Alert) {
	if s.Next != nil {
		s.Next.Emit(alert)
	}
	if s.Store == nil {
		return
	}
	if err := s.Store.SaveSecurityAlert(alert.toDomain()); err != nil {
		log.Printf("[security-monitor] failed to persist alert id=%s code=%s: %v", alert.ID, alert.Code, err)
	}
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
	alert := newAlert("incident_reported", "high", "incident", i.OrderID, "negative order scenario registered")
	m.sink.Emit(alert)
	return alert
}

func (m *Monitor) OrderFailed(payload models.OrderResultPayload) Alert {
	alert := newAlert("operator_reported_failure", "high", "operator_response", payload.OrderID, payload.Reason)
	if alert.Message == "" {
		alert.Message = "operator reported failed order result"
	}
	m.sink.Emit(alert)
	return alert
}

func (m *Monitor) InvalidSystemMessage(source, message string) Alert {
	alert := newAlert("invalid_system_message", "medium", source, "", message)
	m.sink.Emit(alert)
	return alert
}

func (m *Monitor) DeadLetterPublished(source, message string) Alert {
	alert := newAlert("dead_letter_published", "medium", source, "", message)
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

	alert := newAlert("security_goals_not_covered", "medium", "operator_response", order.ID, "operator offer does not cover all requested security goals")
	m.sink.Emit(alert)
	return []Alert{alert}
}

func newAlert(code, severity, source, orderID, message string) Alert {
	return Alert{
		ID:        uuid.NewString(),
		Code:      strings.TrimSpace(code),
		Severity:  strings.TrimSpace(severity),
		Source:    strings.TrimSpace(source),
		OrderID:   strings.TrimSpace(orderID),
		Message:   strings.TrimSpace(message),
		Status:    "open",
		CreatedAt: time.Now().UTC(),
	}
}

func (a Alert) toDomain() *domain.SecurityAlert {
	return &domain.SecurityAlert{
		ID:        a.ID,
		Code:      a.Code,
		Severity:  a.Severity,
		Source:    a.Source,
		OrderID:   a.OrderID,
		Message:   a.Message,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
	}
}
