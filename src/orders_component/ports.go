package orders_component

import "github.com/kirilltahmazidi/aggregator/src/shared/domain"

type Store interface {
	GetCustomer(id string) (*domain.Customer, bool)
	SaveOrder(o *domain.Order) error
	GetOrder(id string) (*domain.Order, bool)
	ListOrders() []*domain.Order
	ListOrdersByCustomer(customerID string) []*domain.Order
	UpdateOrderStatus(id string, status domain.OrderStatus) bool
}
