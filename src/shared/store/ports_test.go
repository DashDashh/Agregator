package store_test

import (
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/contracts_component"
	"github.com/kirilltahmazidi/aggregator/src/operator_exchange_component"
	"github.com/kirilltahmazidi/aggregator/src/orders_component"
	"github.com/kirilltahmazidi/aggregator/src/registry_component"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

func TestStoreImplementsDomainPorts(t *testing.T) {
	var _ registry_component.Store = (*store.Store)(nil)
	var _ orders_component.Store = (*store.Store)(nil)
	var _ contracts_component.Store = (*store.Store)(nil)
	var _ operator_exchange_component.Store = (*store.Store)(nil)
}
