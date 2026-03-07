package models

import "encoding/json"

// MessageType — тип операции (ОФ1–ОФ9 из ТЗ С6)
type MessageType string

const (
	// ОФ1. Регистрация эксплуатанта дронов
	MsgRegisterOperator MessageType = "register_operator"
	// ОФ2. Регистрация заказчика
	MsgRegisterCustomer MessageType = "register_customer"
	// ОФ3. Создание заказа
	MsgCreateOrder MessageType = "create_order"
	// ОФ4. Ручной поиск и выбор исполнителя
	MsgSelectExecutor MessageType = "select_executor"
	// ОФ5. Автоматизированный поиск исполнителя по критериям
	MsgAutoSearchExecutor MessageType = "auto_search_executor"
	// ОФ6. Заключение умного контракта
	MsgConcludeContract MessageType = "conclude_contract"
	// ОФ7. Подтверждение выполнения контракта заказчиком
	MsgConfirmExecution MessageType = "confirm_execution"
	// ОФ8. Создание спора и автовыплата страховки
	MsgCreateDispute MessageType = "create_dispute"
	// ОФ9. Аналитика и визуализация операционной деятельности
	MsgGetAnalytics MessageType = "get_analytics"

	// Сообщения от эксплуатанта агрегатору (читаем из operator.responses)

	// MsgPriceOffer — эксплуатант сообщает цену, за которую готов выполнить заказ
	MsgPriceOffer MessageType = "price_offer"
	// MsgOrderResult — эксплуатант сообщает о завершении (успех/провал) заказа
	MsgOrderResult MessageType = "order_result"

	// Сообщения от агрегатора  эксплуатанту (пишем в operator.requests)

	// MsgConfirmPrice — пользователь подтвердил цену эксплуатанта и готов работать с ним
	MsgConfirmPrice MessageType = "confirm_price"
)

// Входящий конверт
// Request — сообщение, которое агрегатор читает из топика aggregator.requests
type Request struct {
	RequestID string          `json:"request_id"` //уникальный ID
	Type      MessageType     `json:"type"`       //тип операции
	Payload   json.RawMessage `json:"payload"`    //сами данные
}

// Исходящий конверт

type ResponseStatus string

const (
	StatusOK    ResponseStatus = "ok"
	StatusError ResponseStatus = "error"
)

// Response — сообщение, которое агрегатор пишет в топик aggregator.responses
type Response struct {
	RequestID string         `json:"request_id"`        // тот же ID
	Type      MessageType    `json:"type"`              // тот же тип
	Status    ResponseStatus `json:"status"`            // ok или error
	Payload   interface{}    `json:"payload,omitempty"` //результат
	Error     string         `json:"error,omitempty"`   // текст ошибки, если что-то пошло не так
}

// Payload-модели ОФ1

type RegisterOperatorRequest struct {
	Name    string `json:"name"`
	License string `json:"license"`
	Email   string `json:"email"`
}

type RegisterOperatorResponse struct {
	OperatorID string `json:"operator_id"`
	Message    string `json:"message"`
}

// ОФ2

type RegisterCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type RegisterCustomerResponse struct {
	CustomerID string `json:"customer_id"`
	Message    string `json:"message"`
}

// ОФ3

type CreateOrderRequest struct {
	CustomerID  string  `json:"customer_id"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget"`
	FromLat     float64 `json:"from_lat"`
	FromLon     float64 `json:"from_lon"`
	ToLat       float64 `json:"to_lat"`
	ToLon       float64 `json:"to_lon"`
}

type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ОФ4

type SelectExecutorRequest struct {
	OrderID    string `json:"order_id"`
	OperatorID string `json:"operator_id"`
}

type SelectExecutorResponse struct {
	OrderID    string `json:"order_id"`
	OperatorID string `json:"operator_id"`
	Status     string `json:"status"`
}

// ОФ5

type AutoSearchExecutorRequest struct {
	OrderID   string   `json:"order_id"`
	Criteria  []string `json:"criteria"` // ["cost", "safety", "guarantee"]
	MaxBudget float64  `json:"max_budget"`
}

type Candidate struct {
	OperatorID string  `json:"operator_id"`
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	Price      float64 `json:"price"`
}

type AutoSearchExecutorResponse struct {
	OrderID    string      `json:"order_id"`
	Candidates []Candidate `json:"candidates"`
}

// ОФ6

type ConcludeContractRequest struct {
	OrderID    string  `json:"order_id"`
	OperatorID string  `json:"operator_id"`
	Price      float64 `json:"price"`
}

type ConcludeContractResponse struct {
	ContractID string `json:"contract_id"`
	OrderID    string `json:"order_id"`
	Status     string `json:"status"`
}

// ОФ7

type ConfirmExecutionRequest struct {
	ContractID string `json:"contract_id"`
	CustomerID string `json:"customer_id"`
}

type ConfirmExecutionResponse struct {
	ContractID string `json:"contract_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

// ОФ8

type CreateDisputeRequest struct {
	ContractID  string  `json:"contract_id"`
	CustomerID  string  `json:"customer_id"`
	Description string  `json:"description"`
	ClaimAmount float64 `json:"claim_amount"`
}

type CreateDisputeResponse struct {
	DisputeID       string  `json:"dispute_id"`
	ContractID      string  `json:"contract_id"`
	Status          string  `json:"status"`
	InsurancePayout float64 `json:"insurance_payout"`
	Message         string  `json:"message"`
}

// ОФ9

type GetAnalyticsRequest struct {
	From string `json:"from"` // RFC3339
	To   string `json:"to"`   // RFC3339
}

type GetAnalyticsResponse struct {
	TotalOrders     int     `json:"total_orders"`
	CompletedOrders int     `json:"completed_orders"`
	ActiveContracts int     `json:"active_contracts"`
	TotalRevenue    float64 `json:"total_revenue"`
	Disputes        int     `json:"disputes"`
}

// Сообщения от эксплуатанта агрегатору

// PriceOfferPayload — эксплуатант даёт оферту цены на выполнение заказа.
// Агрегатор сохраняет эту цену в БД и показывает пользователю через GET /orders/{id}.
type PriceOfferPayload struct {
	OrderID          string  `json:"order_id"`
	OperatorID       string  `json:"operator_id"`
	OperatorName     string  `json:"operator_name"`
	Price            float64 `json:"price"`
	EstimatedTimeMin int     `json:"estimated_time_minutes"`
}

// OrderResultPayload — эксплуатант сообщает о результате выполнения заказа.
// Статус заказа обновляется автоматически: Success=true → "completed", Success=false → "dispute".
type OrderResultPayload struct {
	OrderID    string `json:"order_id"`
	OperatorID string `json:"operator_id"`
	Success    bool   `json:"success"`
	Reason     string `json:"reason"` // пустая строка при успехе, описание причины при срыве
}

// Сообщения от агрегатора  эксплуатанту (operator.requests)

// ConfirmPricePayload — пользователь подтвердил цену эксплуатанта.
// Агрегатор пересылает это сообщение эксплуатанту, который ожидает подтверждения перед началом выполнения.
type ConfirmPricePayload struct {
	OrderID       string  `json:"order_id"`
	OperatorID    string  `json:"operator_id"`
	AcceptedPrice float64 `json:"accepted_price"`
}
