package contracts_component

import "github.com/kirilltahmazidi/aggregator/src/shared/domain"

type Store interface {
	GetOrder(id string) (*domain.Order, bool)
	ConfirmPrice(id, operatorID string, acceptedPrice, commissionAmount float64) bool
	ConfirmCompletion(id string) bool
	SetOperatorOffer(orderID, operatorID string, price float64) bool
}
