package registry_component

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/internal/models"
)

const Topic = "components.agregator.registry"

var Actions = []models.MessageType{
	models.MsgRegisterOperator,
	models.MsgRegisterCustomer,
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
	case models.MsgRegisterOperator:
		return h.registerOperator(req), true
	case models.MsgRegisterCustomer:
		return h.registerCustomer(req), true
	default:
		return models.Response{}, false
	}
}

func (h *Handler) registerOperator(req models.Request) models.Response {
	var payload models.RegisterOperatorRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}

	return okResponse(req, models.RegisterOperatorResponse{
		OperatorID: uuid.NewString(),
		Message:    fmt.Sprintf("operator '%s' registered (stub)", payload.Name),
	})
}

func (h *Handler) registerCustomer(req models.Request) models.Response {
	var payload models.RegisterCustomerRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}

	return okResponse(req, models.RegisterCustomerResponse{
		CustomerID: uuid.NewString(),
		Message:    fmt.Sprintf("customer '%s' registered (stub)", payload.Name),
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
	log.Printf("[registry_component] error correlation_id=%s: %s", req.GetCorrelationID(), msg)
	return models.Response{
		Action:        models.ResponseAction,
		Sender:        models.DefaultSender,
		CorrelationID: req.GetCorrelationID(),
		Success:       false,
		Error:         msg,
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
	}
}
