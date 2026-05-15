package publisher

import (
	"context"
	"errors"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type fakeBackend struct {
	orderErr        error
	confirmErr      error
	publishedOrders int
	publishedPrices int
}

func (f *fakeBackend) PublishOrder(context.Context, *store.Order) error {
	f.publishedOrders++
	return f.orderErr
}

func (f *fakeBackend) PublishConfirmPrice(context.Context, models.ConfirmPricePayload) error {
	f.publishedPrices++
	return f.confirmErr
}

func TestMultiPublisherSucceedsWhenAnyBackendSucceeds(t *testing.T) {
	failing := &fakeBackend{orderErr: errors.New("boom")}
	working := &fakeBackend{}

	err := NewMultiPublisher(failing, working).PublishOrder(context.Background(), &store.Order{ID: "order-1"})
	if err != nil {
		t.Fatalf("PublishOrder returned error: %v", err)
	}
	if failing.publishedOrders != 1 || working.publishedOrders != 1 {
		t.Fatalf("published counts = %d/%d, want 1/1", failing.publishedOrders, working.publishedOrders)
	}
}

func TestMultiPublisherFailsWithoutBackends(t *testing.T) {
	err := NewMultiPublisher().PublishConfirmPrice(context.Background(), models.ConfirmPricePayload{OrderID: "order-1"})
	if err == nil {
		t.Fatal("PublishConfirmPrice succeeded without backends")
	}
}
