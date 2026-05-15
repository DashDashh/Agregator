package operator_exchange_component

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	securitymonitor "github.com/kirilltahmazidi/aggregator/src/security_monitor_component"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

type ProcessResult string

const (
	ProcessApplied        ProcessResult = "applied"
	ProcessIgnored        ProcessResult = "ignored"
	ProcessInvalidMessage ProcessResult = "invalid_message"
	ProcessInvalidPayload ProcessResult = "invalid_payload"
	ProcessUnknownAction  ProcessResult = "unknown_action"
)

func ProcessOperatorMessage(store Store, data []byte) (ProcessResult, error) {
	return ProcessOperatorMessageWithMonitor(store, securitymonitor.New(nil), data)
}

func ProcessOperatorMessageWithMonitor(store Store, monitor *securitymonitor.Monitor, data []byte) (ProcessResult, error) {
	if monitor == nil {
		monitor = securitymonitor.New(nil)
	}

	var req models.Request
	if err := json.Unmarshal(data, &req); err != nil {
		return ProcessInvalidMessage, fmt.Errorf("разбор operator message: %w", err)
	}

	switch req.Action {
	case models.MsgPriceOffer:
		var p models.PriceOfferPayload
		if err := json.Unmarshal(req.Payload, &p); err != nil {
			return ProcessInvalidPayload, fmt.Errorf("разбор price_offer: %w", err)
		}
		if order, ok := store.GetOrder(p.OrderID); ok {
			monitor.PriceOfferReceived(order, p)
		}
		if store.SetOperatorOffer(p.OrderID, p.OperatorID, p.Price) {
			return ProcessApplied, nil
		}
		return ProcessIgnored, nil

	case models.MsgOrderResult:
		var p models.OrderResultPayload
		if err := json.Unmarshal(req.Payload, &p); err != nil {
			return ProcessInvalidPayload, fmt.Errorf("разбор order_result: %w", err)
		}
		if !p.Success {
			monitor.OrderFailed(p)
		}
		if store.ProcessOrderResult(p.OrderID, p.Success) {
			return ProcessApplied, nil
		}
		return ProcessIgnored, nil

	case models.MsgIncidentReported:
		var p models.IncidentReportedPayload
		if err := json.Unmarshal(req.Payload, &p); err != nil {
			return ProcessInvalidPayload, fmt.Errorf("разбор incident_reported: %w", err)
		}
		if strings.TrimSpace(p.OrderID) == "" || strings.TrimSpace(p.Reason) == "" || p.DamageAmount < 0 {
			return ProcessInvalidPayload, fmt.Errorf("incident_reported requires order_id, reason and non-negative damage_amount")
		}
		if order, ok := store.GetOrder(p.OrderID); ok && p.OperatorID == "" {
			p.OperatorID = order.OperatorID
		}
		incident := &domain.Incident{
			ID:           uuid.NewString(),
			OrderID:      p.OrderID,
			OperatorID:   p.OperatorID,
			ReporterID:   p.ReporterID,
			Reason:       strings.TrimSpace(p.Reason),
			Description:  strings.TrimSpace(p.Description),
			DamageAmount: p.DamageAmount,
			Status:       "registered",
			CreatedAt:    time.Now().UTC(),
		}
		if err := store.RegisterIncident(incident); err != nil {
			return ProcessIgnored, fmt.Errorf("регистрация инцидента: %w", err)
		}
		monitor.IncidentReported(*incident)
		return ProcessApplied, nil

	default:
		return ProcessUnknownAction, nil
	}
}
