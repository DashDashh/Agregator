package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/internal/models"
)

// Пока что это все заглушки
type Handler struct{}

func New() *Handler {
	return &Handler{}
}

// Handle — основная точка диспетчеризации: по типу сообщения вызывает нужный обработчик
func (h *Handler) Handle(req models.Request) models.Response {
	log.Printf("[handler] processing correlation_id=%s action=%s", req.GetCorrelationID(), req.Action)

	switch req.Action {
	case models.MsgRegisterOperator:
		return h.registerOperator(req)
	case models.MsgRegisterCustomer:
		return h.registerCustomer(req)
	case models.MsgCreateOrder:
		return h.createOrder(req)
	case models.MsgSelectExecutor:
		return h.selectExecutor(req)
	case models.MsgAutoSearchExecutor:
		return h.autoSearchExecutor(req)
	case models.MsgConcludeContract:
		return h.concludeContract(req)
	case models.MsgConfirmExecution:
		return h.confirmExecution(req)
	case models.MsgCreateDispute:
		return h.createDispute(req)
	case models.MsgGetAnalytics:
		return h.getAnalytics(req)
	default:
		return errResponse(req, fmt.Sprintf("unknown action: %s", req.Action))
	}
}

// ОФ1 Регистрация эксплуатанта дронов

func (h *Handler) registerOperator(req models.Request) models.Response {
	var p models.RegisterOperatorRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	// заглушка: генерируем ID, сохранение в БД — позже
	return okResponse(req, models.RegisterOperatorResponse{
		OperatorID: uuid.NewString(),
		Message:    fmt.Sprintf("operator '%s' registered (stub)", p.Name),
	})
}

// ОФ2 Регистрация заказчика

func (h *Handler) registerCustomer(req models.Request) models.Response {
	var p models.RegisterCustomerRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	// заглушка
	return okResponse(req, models.RegisterCustomerResponse{
		CustomerID: uuid.NewString(),
		Message:    fmt.Sprintf("customer '%s' registered (stub)", p.Name),
	})
}

// ОФ3 Создание заказа

func (h *Handler) createOrder(req models.Request) models.Response {
	var p models.CreateOrderRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	// заглушка
	return okResponse(req, models.CreateOrderResponse{
		OrderID: uuid.NewString(),
		Status:  "pending",
		Message: "order created, awaiting executor selection (stub)",
	})
}

// ОФ4 Ручной поиск и выбор исполнителя

func (h *Handler) selectExecutor(req models.Request) models.Response {
	var p models.SelectExecutorRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	// заглушка
	return okResponse(req, models.SelectExecutorResponse{
		OrderID:    p.OrderID,
		OperatorID: p.OperatorID,
		Status:     "executor_selected",
	})
}

// ОФ5 Автоматизированный поиск исполнителя

func (h *Handler) autoSearchExecutor(req models.Request) models.Response {
	var p models.AutoSearchExecutorRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	// заглушка: возвращаем одного кандидата с фиксированным score
	candidates := []models.Candidate{
		{
			OperatorID: uuid.NewString(),
			Name:       "Stub Operator Alpha",
			Score:      0.95,
			Price:      p.MaxBudget * 0.8,
		},
		{
			OperatorID: uuid.NewString(),
			Name:       "Stub Operator Beta",
			Score:      0.87,
			Price:      p.MaxBudget * 0.6,
		},
	}
	return okResponse(req, models.AutoSearchExecutorResponse{
		OrderID:    p.OrderID,
		Candidates: candidates,
	})
}

// ОФ6 Заключение умного контракта

func (h *Handler) concludeContract(req models.Request) models.Response {
	var p models.ConcludeContractRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	// заглшука
	return okResponse(req, models.ConcludeContractResponse{
		ContractID: uuid.NewString(),
		OrderID:    p.OrderID,
		Status:     "active",
	})
}

// ОФ7 Подтверждение выполнения контракта заказчиком

func (h *Handler) confirmExecution(req models.Request) models.Response {
	var p models.ConfirmExecutionRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	// заглушка
	return okResponse(req, models.ConfirmExecutionResponse{
		ContractID: p.ContractID,
		Status:     "completed",
		Message:    "contract marked as completed by customer (stub)",
	})
}

// ОФ8 Создание спора и автоматическая выплата страховки

func (h *Handler) createDispute(req models.Request) models.Response {
	var p models.CreateDisputeRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	// заглушка: всегда одобряем и выплачиваем 100% заявленной суммы
	return okResponse(req, models.CreateDisputeResponse{
		DisputeID:       uuid.NewString(),
		ContractID:      p.ContractID,
		Status:          "dispute_opened",
		InsurancePayout: p.ClaimAmount,
		Message:         "dispute registered, insurance payout initiated (stub)",
	})
}

// ОФ9 Аналитика и визуализация

func (h *Handler) getAnalytics(req models.Request) models.Response {
	var p models.GetAnalyticsRequest
	if err := json.Unmarshal(req.Payload, &p); err != nil {
		return errResponse(req, "invalid payload: "+err.Error())
	}
	_ = p // будет использоваться при реальной реализации
	// заглушка
	return okResponse(req, models.GetAnalyticsResponse{
		TotalOrders:     42,
		CompletedOrders: 38,
		ActiveContracts: 4,
		TotalRevenue:    125000.00,
		Disputes:        2,
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
	log.Printf("[handler] error correlation_id=%s: %s", req.GetCorrelationID(), msg)
	return models.Response{
		Action:        models.ResponseAction,
		Sender:        models.DefaultSender,
		CorrelationID: req.GetCorrelationID(),
		Success:       false,
		Error:         msg,
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
	}
}
