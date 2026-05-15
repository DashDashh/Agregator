package gateway

import (
	"fmt"

	"github.com/kirilltahmazidi/aggregator/src/analytics_component"
	"github.com/kirilltahmazidi/aggregator/src/contracts_component"
	"github.com/kirilltahmazidi/aggregator/src/gateway/bus/handler"
	"github.com/kirilltahmazidi/aggregator/src/orders_component"
	"github.com/kirilltahmazidi/aggregator/src/registry_component"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
)

const (
	ComponentRegistry  = registry_component.Topic
	ComponentOrders    = orders_component.Topic
	ComponentContracts = contracts_component.Topic
	ComponentAnalytics = analytics_component.Topic
)

// ActionRouting описывает, какой компонент отвечает за action.
// В режиме inprocess gateway вызывает обработчик напрямую, в режиме broker
// использует эту же таблицу для отправки запроса в топик отдельного сервиса.
var ActionRouting = map[models.MessageType]string{
	models.MsgRegisterOperator:   ComponentRegistry,
	models.MsgRegisterCustomer:   ComponentRegistry,
	models.MsgCreateOrder:        ComponentOrders,
	models.MsgSelectExecutor:     ComponentOrders,
	models.MsgAutoSearchExecutor: ComponentOrders,
	models.MsgConcludeContract:   ComponentContracts,
	models.MsgConfirmExecution:   ComponentContracts,
	models.MsgCreateDispute:      ComponentContracts,
	models.MsgGetAnalytics:       ComponentAnalytics,
}

type Gateway struct {
	handler *handler.Handler
}

func New(h *handler.Handler) *Gateway {
	return &Gateway{handler: h}
}

func (g *Gateway) Route(req models.Request) models.Response {
	if _, ok := ActionRouting[req.Action]; !ok {
		return models.Response{
			Action:        models.ResponseAction,
			Sender:        models.DefaultSender,
			CorrelationID: req.GetCorrelationID(),
			Success:       false,
			Error:         fmt.Sprintf("unknown action: %s", req.Action),
			Timestamp:     req.Timestamp,
		}
	}
	return g.handler.Handle(req)
}

func (g *Gateway) ComponentFor(action models.MessageType) (string, bool) {
	component, ok := ActionRouting[action]
	return component, ok
}
