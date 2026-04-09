package contracts_component

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/internal/models"
	"github.com/kirilltahmazidi/aggregator/internal/response"
)

const Topic = "components.agregator.contracts"

var Actions = []models.MessageType{
	models.MsgConcludeContract,
	models.MsgConfirmExecution,
	models.MsgCreateDispute,
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
	case models.MsgConcludeContract:
		return h.concludeContract(req), true
	case models.MsgConfirmExecution:
		return h.confirmExecution(req), true
	case models.MsgCreateDispute:
		return h.createDispute(req), true
	default:
		return models.Response{}, false
	}
}

func (h *Handler) concludeContract(req models.Request) models.Response {
	var payload models.ConcludeContractRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}

	return okResponse(req, models.ConcludeContractResponse{
		ContractID: uuid.NewString(),
		OrderID:    payload.OrderID,
		Status:     "active",
	})
}

func (h *Handler) confirmExecution(req models.Request) models.Response {
	var payload models.ConfirmExecutionRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}

	return okResponse(req, models.ConfirmExecutionResponse{
		ContractID: payload.ContractID,
		Status:     "completed",
		Message:    "contract marked as completed by customer (stub)",
	})
}

func (h *Handler) createDispute(req models.Request) models.Response {
	var payload models.CreateDisputeRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}

	return okResponse(req, models.CreateDisputeResponse{
		DisputeID:       uuid.NewString(),
		ContractID:      payload.ContractID,
		Status:          "dispute_opened",
		InsurancePayout: payload.ClaimAmount,
		Message:         "dispute registered, insurance payout initiated (stub)",
	})
}

func okResponse(req models.Request, payload interface{}) models.Response {
	return response.OK(req, payload)
}

func errResponse(req models.Request, msg string) models.Response {
	log.Printf("[contracts_component] error correlation_id=%s: %s", req.GetCorrelationID(), msg)
	return response.Err(req, msg)
}
