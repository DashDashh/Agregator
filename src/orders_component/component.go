package orders_component

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/internal/models"
)

const Topic = "components.agregator.orders"

var Actions = []models.MessageType{
	models.MsgCreateOrder,
	models.MsgSelectExecutor,
	models.MsgAutoSearchExecutor,
}

func Handles(action models.MessageType) bool {
	for _, candidate := range Actions {
		if candidate == action {
			return true
		}
	}
	return false
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(req models.Request) (models.Response, bool) {
	switch req.Action {
	case models.MsgCreateOrder:
		return h.createOrder(req), true
	case models.MsgSelectExecutor:
		return h.selectExecutor(req), true
	case models.MsgAutoSearchExecutor:
		return h.autoSearchExecutor(req), true
	default:
		return models.Response{}, false
	}
}

func (h *Handler) createOrder(req models.Request) models.Response {
	var payload models.CreateOrderRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}

	return okResponse(req, models.CreateOrderResponse{
		OrderID: uuid.NewString(),
		Status:  "pending",
		Message: "order created, awaiting executor selection (stub)",
	})
}

func (h *Handler) selectExecutor(req models.Request) models.Response {
	var payload models.SelectExecutorRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}

	return okResponse(req, models.SelectExecutorResponse{
		OrderID:    payload.OrderID,
		OperatorID: payload.OperatorID,
		Status:     "executor_selected",
	})
}

func (h *Handler) autoSearchExecutor(req models.Request) models.Response {
	var payload models.AutoSearchExecutorRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}

	return okResponse(req, models.AutoSearchExecutorResponse{
		OrderID: payload.OrderID,
		Candidates: []models.Candidate{
			{
				OperatorID: uuid.NewString(),
				Name:       "Stub Operator Alpha",
				Score:      0.95,
				Price:      payload.MaxBudget * 0.8,
			},
			{
				OperatorID: uuid.NewString(),
				Name:       "Stub Operator Beta",
				Score:      0.87,
				Price:      payload.MaxBudget * 0.6,
			},
		},
	})
}

func okResponse(req models.Request, payload interface{}) models.Response {
	return models.Response{
		Action:        models.ResponseAction,
		Payload:       payload,
		Sender:        models.DefaultSender,
		CorrelationID: req.GetCorrelationID(),
		Success:       true,
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func errResponse(req models.Request, msg string) models.Response {
	log.Printf("[orders_component] error correlation_id=%s: %s", req.GetCorrelationID(), msg)
	return models.Response{
		Action:        models.ResponseAction,
		Sender:        models.DefaultSender,
		CorrelationID: req.GetCorrelationID(),
		Success:       false,
		Error:         msg,
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
	}
}
