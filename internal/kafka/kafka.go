package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/kirilltahmazidi/aggregator/internal/config"
	"github.com/kirilltahmazidi/aggregator/internal/handler"
	"github.com/kirilltahmazidi/aggregator/internal/models"
	"github.com/kirilltahmazidi/aggregator/internal/store"
	kafkago "github.com/segmentio/kafka-go"
)

// Service инкапсулирует kafka reader/writer и запускает цикл обработки.
type Service struct {
	reader    *kafkago.Reader
	writer    *kafkago.Writer // пишет responses обратно (для других сервисов)
	outWriter *kafkago.Writer // пишет задания эксплуатантам в operator.requests
	dlt       *kafkago.Writer // dead-letter topic для нечитаемых сообщений
	handler   *handler.Handler
	store     *store.Store // для обновления статусов заказов
}

func NewService(cfg *config.Config, h *handler.Handler, s *store.Store) *Service {
	// читает из aggregator.requests
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  []string{cfg.KafkaBroker},
		Topic:    cfg.RequestTopic,  //откуда читаем
		GroupID:  cfg.ConsumerGroup, // имя группы
		MinBytes: 1,
		MaxBytes: 10e6, // 10 MB
		Logger:   kafkago.LoggerFunc(func(msg string, args ...interface{}) { log.Printf("[kafka/reader] "+msg, args...) }),
	})

	// пишет в aggregator.responses
	writer := &kafkago.Writer{
		Addr:     kafkago.TCP(cfg.KafkaBroker),
		Topic:    cfg.ResponseTopic, //куда пишем
		Balancer: &kafkago.LeastBytes{},
		Logger:   kafkago.LoggerFunc(func(msg string, args ...interface{}) { log.Printf("[kafka/writer] "+msg, args...) }),
	}

	// пишет задания эксплуатантам в operator.requests
	outWriter := &kafkago.Writer{
		Addr:     kafkago.TCP(cfg.KafkaBroker),
		Topic:    cfg.OperatorTopic,
		Balancer: &kafkago.LeastBytes{},
		Logger:   kafkago.LoggerFunc(func(msg string, args ...interface{}) { log.Printf("[kafka/out] "+msg, args...) }),
	}

	// это кароче если пришел мусор который нельзя прочитать => кладем в отдельный топик
	dlt := &kafkago.Writer{
		Addr:     kafkago.TCP(cfg.KafkaBroker),
		Topic:    cfg.DeadLetterTopic,
		Balancer: &kafkago.LeastBytes{},
	}

	return &Service{
		reader:    reader,
		writer:    writer,
		outWriter: outWriter,
		dlt:       dlt,
		handler:   h,
		store:     s,
	}
}

// Run запускает бесконечный цикл чтения из топика запросов.
// Завершается при отмене ctx.
// PublishOrder отправляет заказ в топик operator.requests — эксплуатанты его читают.
// Вызывается из HTTP-обработчика когда фронт создаёт заказ.
func (s *Service) PublishOrder(ctx context.Context, order *store.Order) error {
	// собираем payload в формате models.CreateOrderRequest
	payload, err := json.Marshal(models.CreateOrderRequest{
		CustomerID:  order.CustomerID,
		Description: order.Description,
		Budget:      order.Budget,
		FromLat:     order.FromLat,
		FromLon:     order.FromLon,
		ToLat:       order.ToLat,
		ToLon:       order.ToLon,
	})
	if err != nil {
		return err
	}

	// оборачиваем в стандартный конверт Request — чтобы формат совпадал с остальными сообщениями
	req := models.Request{
		RequestID: order.ID, // используем ID заказа как ID запроса — связь между HTTP и Kafka
		Type:      models.MsgCreateOrder,
		Payload:   payload,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	err = s.outWriter.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(order.ID), // ключ = ID заказа, Kafka использует его для партиционирования
		Value: data,
	})
	if err != nil {
		return err
	}

	log.Printf("[kafka] order published to operators order_id=%s", order.ID)
	return nil
}

func (s *Service) Run(ctx context.Context) error {
	log.Printf("[kafka] starting consumer loop on topic=%s", s.reader.Config().Topic)
	defer s.reader.Close()
	defer s.writer.Close()
	defer s.outWriter.Close()
	defer s.dlt.Close()

	for {
		select {
		case <-ctx.Done():
			log.Println("[kafka] context cancelled, shutting down consumer")
			return ctx.Err()
		default:
		}

		msg, err := s.reader.ReadMessage(ctx) // ждем сообщение
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("[kafka] read error: %v", err)
			continue
		}

		log.Printf("[kafka] received message offset=%d partition=%d key=%s",
			msg.Offset, msg.Partition, string(msg.Key))

		s.processMessage(ctx, msg) // обрабатываем
	}
}

func (s *Service) processMessage(ctx context.Context, msg kafkago.Message) {
	var req models.Request
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		log.Printf("[kafka] cannot unmarshal message: %v — sending to DLT", err)
		s.sendToDLT(ctx, msg)
		return
	}

	resp := s.handler.Handle(req)

	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[kafka] cannot marshal response for request_id=%s: %v", req.RequestID, err)
		return
	}

	out := kafkago.Message{
		Key:   []byte(req.RequestID),
		Value: respBytes,
	}

	if err := s.writer.WriteMessages(ctx, out); err != nil {
		log.Printf("[kafka] failed to write response for request_id=%s: %v", req.RequestID, err)
	} else {
		log.Printf("[kafka] response sent for request_id=%s status=%s", req.RequestID, resp.Status)
	}

	// Обновляем статус заказа в store — фронт увидит изменение через GET /orders/{id}
	if req.Type == models.MsgCreateOrder && resp.Status == models.StatusOK {
		if s.store.UpdateOrderStatus(req.RequestID, store.StatusSearching) {
			log.Printf("[kafka] order status updated to searching order_id=%s", req.RequestID)
		}
	}
}

func (s *Service) sendToDLT(ctx context.Context, original kafkago.Message) {
	if err := s.dlt.WriteMessages(ctx, kafkago.Message{
		Key:   original.Key,
		Value: original.Value,
	}); err != nil {
		log.Printf("[kafka] failed to write to DLT: %v", err)
	}
}
