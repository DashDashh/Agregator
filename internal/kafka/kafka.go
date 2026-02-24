package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/kirilltahmazidi/aggregator/internal/config"
	"github.com/kirilltahmazidi/aggregator/internal/handler"
	"github.com/kirilltahmazidi/aggregator/internal/models"
	kafkago "github.com/segmentio/kafka-go"
)

// Service инкапсулирует kafka reader/writer и запускает цикл обработки.
type Service struct {
	reader  *kafkago.Reader
	writer  *kafkago.Writer
	dlt     *kafkago.Writer // dead-letter topic для нечитаемых сообщений
	handler *handler.Handler
}

func NewService(cfg *config.Config, h *handler.Handler) *Service {
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

	// это кароче если пришел мусор который нельзя прочитать => кладем в отдельный топик
	dlt := &kafkago.Writer{
		Addr:     kafkago.TCP(cfg.KafkaBroker),
		Topic:    cfg.DeadLetterTopic,
		Balancer: &kafkago.LeastBytes{},
	}

	return &Service{
		reader:  reader,
		writer:  writer,
		dlt:     dlt,
		handler: h,
	}
}

// Run запускает бесконечный цикл чтения из топика запросов.
// Завершается при отмене ctx.
func (s *Service) Run(ctx context.Context) error {
	log.Printf("[kafka] starting consumer loop on topic=%s", s.reader.Config().Topic)
	defer s.reader.Close()
	defer s.writer.Close()
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
}

func (s *Service) sendToDLT(ctx context.Context, original kafkago.Message) {
	if err := s.dlt.WriteMessages(ctx, kafkago.Message{
		Key:   original.Key,
		Value: original.Value,
	}); err != nil {
		log.Printf("[kafka] failed to write to DLT: %v", err)
	}
}
