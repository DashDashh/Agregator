package contracts_component

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	"github.com/kirilltahmazidi/aggregator/src/shared/response"
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

type Handler struct {
	store          Store
	commissionRate float64
}

func NewHandler() *Handler {
	return &Handler{}
}

func NewStoreHandler(store Store, commissionRate float64) *Handler {
	return &Handler{store: store, commissionRate: commissionRate}
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
		return response.Err("contracts_component", req, "invalid payload: "+err.Error())
	}
	if h.store != nil {
		if payload.OrderID == "" || payload.OperatorID == "" || payload.Price <= 0 {
			return response.Err("contracts_component", req, "order_id, operator_id and positive price are required")
		}
		commission := payload.Price * h.commissionRate
		if !h.store.ConfirmPrice(payload.OrderID, payload.OperatorID, payload.Price, commission) {
			return response.Err("contracts_component", req, "cannot conclude contract for current order state")
		}
	}

	return response.Ok(req, models.ConcludeContractResponse{
		ContractID: payload.OrderID,
		OrderID:    payload.OrderID,
		Status:     "active",
	})
}

func (h *Handler) confirmExecution(req models.Request) models.Response {
	var payload models.ConfirmExecutionRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return response.Err("contracts_component", req, "invalid payload: "+err.Error())
	}
	if h.store != nil {
		if payload.ContractID == "" {
			return response.Err("contracts_component", req, "contract_id is required")
		}
		if !h.store.ConfirmCompletion(payload.ContractID) {
			return response.Err("contracts_component", req, "cannot confirm execution for current order state")
		}
	}

	return response.Ok(req, models.ConfirmExecutionResponse{
		ContractID: payload.ContractID,
		Status:     "completed",
		Message:    "contract marked as completed by customer (stub)",
	})
}

func (h *Handler) createDispute(req models.Request) models.Response {
	var payload models.CreateDisputeRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return response.Err("contracts_component", req, "invalid payload: "+err.Error())
	}
	if h.store != nil {
		if payload.ContractID == "" || payload.Description == "" {
			return response.Err("contracts_component", req, "contract_id and description are required")
		}
		order, ok := h.store.GetOrder(payload.ContractID)
		if !ok {
			return response.Err("contracts_component", req, "order not found")
		}
		incident := &domain.Incident{
			ID:           uuid.NewString(),
			OrderID:      payload.ContractID,
			OperatorID:   order.OperatorID,
			ReporterID:   payload.CustomerID,
			Reason:       "dispute",
			Description:  payload.Description,
			DamageAmount: payload.ClaimAmount,
			Status:       "registered",
			CreatedAt:    time.Now().UTC(),
		}
		if err := h.store.RegisterIncident(incident); err != nil {
			return response.Err("contracts_component", req, "register incident: "+err.Error())
		}
	}

	return response.Ok(req, models.CreateDisputeResponse{
		DisputeID:       uuid.NewString(),
		ContractID:      payload.ContractID,
		Status:          "dispute_opened",
		InsurancePayout: payload.ClaimAmount,
		Message:         "dispute registered, insurance payout initiated (stub)",
	})
}
