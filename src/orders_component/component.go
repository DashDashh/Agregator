package orders_component

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	"github.com/kirilltahmazidi/aggregator/src/shared/response"
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
		return response.Err("orders_component", req, "invalid payload: "+err.Error())
	}
	if h.store == nil {
		return response.Ok(req, models.CreateOrderResponse{
			OrderID: uuid.NewString(),
			Status:  "pending",
			Message: "order created, awaiting executor selection (stub)",
		})
	}
	if payload.CustomerID == "" || payload.Description == "" {
		return response.Err("orders_component", req, "customer_id and description are required")
	}
	if _, ok := h.store.GetCustomer(payload.CustomerID); !ok {
		return response.Err("orders_component", req, "customer not found")
	}
	orderID := req.GetCorrelationID()
	if orderID == "" {
		orderID = uuid.NewString()
	}
	missionType := payload.MissionType
	if missionType == "" {
		missionType = "delivery"
	}
	order := &domain.Order{
		ID:             orderID,
		CustomerID:     payload.CustomerID,
		Description:    payload.Description,
		Budget:         payload.Budget,
		FromLat:        payload.FromLat,
		FromLon:        payload.FromLon,
		ToLat:          payload.ToLat,
		ToLon:          payload.ToLon,
		MissionType:    missionType,
		SecurityGoals:  payload.SecurityGoals,
		TopLeftLat:     payload.TopLeftLat,
		TopLeftLon:     payload.TopLeftLon,
		BottomRightLat: payload.BottomRightLat,
		BottomRightLon: payload.BottomRightLon,
		Status:         domain.StatusSearching,
		CreatedAt:      time.Now().UTC(),
	}
	if err := h.store.SaveOrder(order); err != nil {
		return response.Err("orders_component", req, "save order: "+err.Error())
	}

	return response.Ok(req, models.CreateOrderResponse{
		OrderID: order.ID,
		Status:  string(order.Status),
		Message: "order created and published for executor search",
	})
}

func (h *Handler) selectExecutor(req models.Request) models.Response {
	var payload models.SelectExecutorRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return response.Err("orders_component", req, "invalid payload: "+err.Error())
	}
	if h.store != nil {
		if payload.OrderID == "" || payload.OperatorID == "" {
			return response.Err("orders_component", req, "order_id and operator_id are required")
		}
		if _, ok := h.store.GetOrder(payload.OrderID); !ok {
			return response.Err("orders_component", req, "order not found")
		}
	}

	return response.Ok(req, models.SelectExecutorResponse{
		OrderID:    payload.OrderID,
		OperatorID: payload.OperatorID,
		Status:     "executor_selected",
	})
}

func (h *Handler) autoSearchExecutor(req models.Request) models.Response {
	var payload models.AutoSearchExecutorRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		return response.Err("orders_component", req, "invalid payload: "+err.Error())
	}
	if h.store != nil {
		if _, ok := h.store.GetOrder(payload.OrderID); !ok {
			return response.Err("orders_component", req, "order not found")
		}
	}

	return response.Ok(req, models.AutoSearchExecutorResponse{
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
