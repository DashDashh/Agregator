package kafka

import (
	"context"
	"encoding/json"
	"testing"

	busgateway "github.com/kirilltahmazidi/aggregator/src/gateway/bus/gateway"
	"github.com/kirilltahmazidi/aggregator/src/gateway/config"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	kafkago "github.com/segmentio/kafka-go"
)

type fakeStore struct {
	order           *domain.Order
	offerOrderID    string
	offerOperatorID string
	offerPrice      float64
}

func (f *fakeStore) GetOrder(id string) (*domain.Order, bool) {
	if f.order != nil && f.order.ID == id {
		return f.order, true
	}
	return nil, false
}

func (f *fakeStore) SetOperatorOffer(orderID, operatorID string, price float64) bool {
	f.offerOrderID = orderID
	f.offerOperatorID = operatorID
	f.offerPrice = price
	return true
}

func (f *fakeStore) ProcessOrderResult(string, bool) bool {
	return true
}

func (f *fakeStore) UpdateOrderStatus(string, domain.OrderStatus) bool {
	return true
}

func (f *fakeStore) RegisterIncident(*domain.Incident) error {
	return nil
}

func TestNewServiceConfiguresKafkaResources(t *testing.T) {
	cfg := &config.Config{
		KafkaBroker:           "localhost:9092",
		RequestTopic:          "requests",
		ResponseTopic:         "responses",
		ConsumerGroup:         "group",
		DeadLetterTopic:       "dlt",
		OperatorTopic:         "operator.requests",
		OperatorResponseTopic: "operator.responses",
		ComponentDispatchMode: "inprocess",
	}

	svc := NewService(cfg, busgateway.New(nil), &fakeStore{})
	defer svc.reader.Close()
	defer svc.writer.Close()
	defer svc.componentWriter.Close()
	defer svc.outWriter.Close()
	defer svc.operatorReader.Close()
	defer svc.dlt.Close()

	if svc.reader.Config().Topic != "requests" {
		t.Fatalf("reader topic = %q", svc.reader.Config().Topic)
	}
	if svc.writer.Topic != "responses" {
		t.Fatalf("writer topic = %q", svc.writer.Topic)
	}
	if svc.outWriter.Topic != "operator.requests" {
		t.Fatalf("out writer topic = %q", svc.outWriter.Topic)
	}
	if svc.operatorReader.Config().Topic != "operator.responses" {
		t.Fatalf("operator reader topic = %q", svc.operatorReader.Config().Topic)
	}
	if svc.dispatchMode != "inprocess" {
		t.Fatalf("dispatchMode = %q", svc.dispatchMode)
	}
}

func TestRunReturnsCancelledContext(t *testing.T) {
	cfg := &config.Config{
		KafkaBroker:           "localhost:9092",
		RequestTopic:          "requests",
		ResponseTopic:         "responses",
		ConsumerGroup:         "group",
		DeadLetterTopic:       "dlt",
		OperatorTopic:         "operator.requests",
		OperatorResponseTopic: "operator.responses",
		ComponentDispatchMode: "inprocess",
	}
	svc := NewService(cfg, busgateway.New(nil), &fakeStore{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := svc.Run(ctx); err != context.Canceled {
		t.Fatalf("Run error = %v, want context.Canceled", err)
	}
}

func TestRunOperatorConsumerReturnsCancelledContext(t *testing.T) {
	cfg := &config.Config{
		KafkaBroker:           "localhost:9092",
		RequestTopic:          "requests",
		ResponseTopic:         "responses",
		ConsumerGroup:         "group",
		DeadLetterTopic:       "dlt",
		OperatorTopic:         "operator.requests",
		OperatorResponseTopic: "operator.responses",
		ComponentDispatchMode: "inprocess",
	}
	svc := NewService(cfg, busgateway.New(nil), &fakeStore{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := svc.RunOperatorConsumer(ctx); err != context.Canceled {
		t.Fatalf("RunOperatorConsumer error = %v, want context.Canceled", err)
	}
}

func TestProcessOperatorMessageUpdatesStore(t *testing.T) {
	payload, err := json.Marshal(models.PriceOfferPayload{
		OrderID:    "order-1",
		OperatorID: "operator-1",
		Price:      1200,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	data, err := json.Marshal(models.Request{Action: models.MsgPriceOffer, Payload: payload})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	store := &fakeStore{}
	svc := &Service{store: store}

	svc.processOperatorMessage(context.Background(), kafkago.Message{Value: data})

	if store.offerOrderID != "order-1" || store.offerOperatorID != "operator-1" || store.offerPrice != 1200 {
		t.Fatalf("stored offer = %s/%s/%v", store.offerOrderID, store.offerOperatorID, store.offerPrice)
	}
}
