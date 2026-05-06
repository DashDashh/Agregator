package operator_exchange_component

import "github.com/kirilltahmazidi/aggregator/src/shared/domain"

type Store interface {
	SetOperatorOffer(orderID, operatorID string, price float64) bool
	ProcessOrderResult(orderID string, success bool) bool
	UpdateOrderStatus(id string, status domain.OrderStatus) bool
}
