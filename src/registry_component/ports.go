package registry_component

import "github.com/kirilltahmazidi/aggregator/src/shared/domain"

type Store interface {
	SaveCustomer(c *domain.Customer) error
	GetCustomer(id string) (*domain.Customer, bool)
	GetCustomerByEmail(email string) (*domain.Customer, bool)
	SetCustomerPasswordHash(id, passwordHash string) bool
	SaveOperator(op *domain.Operator) error
	GetOperator(id string) (*domain.Operator, bool)
	GetOperatorByEmail(email string) (*domain.Operator, bool)
	SetOperatorPasswordHash(id, passwordHash string) bool
}
