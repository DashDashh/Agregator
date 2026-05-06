package operator_exchange_component

import (
	"encoding/json"
	"fmt"

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
		if store.SetOperatorOffer(p.OrderID, p.OperatorID, p.Price) {
			return ProcessApplied, nil
		}
		return ProcessIgnored, nil

	case models.MsgOrderResult:
		var p models.OrderResultPayload
		if err := json.Unmarshal(req.Payload, &p); err != nil {
			return ProcessInvalidPayload, fmt.Errorf("разбор order_result: %w", err)
		}
		if store.ProcessOrderResult(p.OrderID, p.Success) {
			return ProcessApplied, nil
		}
		return ProcessIgnored, nil

	default:
		return ProcessUnknownAction, nil
	}
}
