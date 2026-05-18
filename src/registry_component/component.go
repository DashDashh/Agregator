package registry_component

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	"github.com/kirilltahmazidi/aggregator/src/shared/response"
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

type Handler struct {
	store Store
}

func NewHandler() *Handler {
	return &Handler{}
}

func NewStoreHandler(store Store) *Handler {
	return &Handler{store: store}
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
		return response.Err("registry_component", req, "invalid payload: "+err.Error())
	}
	if h.store == nil {
		return response.Ok(req, models.RegisterOperatorResponse{
			OperatorID: uuid.NewString(),
			Message:    fmt.Sprintf("operator '%s' registered (stub)", payload.Name),
		})
	}
	if payload.Name == "" || payload.License == "" || payload.Email == "" {
		return response.Err("registry_component", req, "name, license and email are required")
	}
	if existing, ok := h.store.GetOperatorByEmail(payload.Email); ok {
		return response.Ok(req, models.RegisterOperatorResponse{
			OperatorID: existing.ID,
			Message:    fmt.Sprintf("operator '%s' already registered", existing.Name),
		})
	}
	passwordHash, err := optionalPasswordHash(payload.Password)
	if err != nil {
		return response.Err("registry_component", req, err.Error())
	}
	op := &domain.Operator{
		ID:           uuid.NewString(),
		Name:         payload.Name,
		License:      payload.License,
		Email:        payload.Email,
		PasswordHash: passwordHash,
	}
	if err := h.store.SaveOperator(op); err != nil {
		return response.Err("registry_component", req, "save operator: "+err.Error())
	}

	return response.Ok(req, models.RegisterOperatorResponse{
		OperatorID: op.ID,
		Message:    fmt.Sprintf("operator '%s' registered", payload.Name),
	})
}

func (h *Handler) registerCustomer(req models.Request) models.Response {
	var payload models.RegisterCustomerRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return response.Err("registry_component", req, "invalid payload: "+err.Error())
	}
	if h.store == nil {
		return response.Ok(req, models.RegisterCustomerResponse{
			CustomerID: uuid.NewString(),
			Message:    fmt.Sprintf("customer '%s' registered (stub)", payload.Name),
		})
	}
	if payload.Name == "" || payload.Email == "" {
		return response.Err("registry_component", req, "name and email are required")
	}
	if existing, ok := h.store.GetCustomerByEmail(payload.Email); ok {
		return response.Ok(req, models.RegisterCustomerResponse{
			CustomerID: existing.ID,
			Message:    fmt.Sprintf("customer '%s' already registered", existing.Name),
		})
	}
	passwordHash, err := optionalPasswordHash(payload.Password)
	if err != nil {
		return response.Err("registry_component", req, err.Error())
	}
	customer := &domain.Customer{
		ID:           uuid.NewString(),
		Name:         payload.Name,
		Email:        payload.Email,
		Phone:        payload.Phone,
		PasswordHash: passwordHash,
	}
	if err := h.store.SaveCustomer(customer); err != nil {
		return response.Err("registry_component", req, "save customer: "+err.Error())
	}

	return response.Ok(req, models.RegisterCustomerResponse{
		CustomerID: customer.ID,
		Message:    fmt.Sprintf("customer '%s' registered", payload.Name),
	})
}

func optionalPasswordHash(password string) (string, error) {
	if password == "" {
		return "", nil
	}
	return auth.HashPassword(password)
}
