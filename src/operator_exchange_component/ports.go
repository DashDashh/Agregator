package operator_exchange_component

import "github.com/kirilltahmazidi/aggregator/src/shared/domain"

type Store interface {
	GetOrder(id string) (*domain.Order, bool)
	SetOperatorOffer(orderID, operatorID string, price float64) bool
	ProcessOrderResult(orderID string, success bool) bool
	UpdateOrderStatus(id string, status domain.OrderStatus) bool
	RegisterIncident(i *domain.Incident) error
}
