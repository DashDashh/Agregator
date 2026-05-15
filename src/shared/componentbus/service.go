package componentbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/kirilltahmazidi/aggregator/src/shared/kafkautil"
	"github.com/kirilltahmazidi/aggregator/src/shared/models"
	kafkago "github.com/segmentio/kafka-go"
)

type Handler interface {
	Handle(req models.Request) (models.Response, bool)
}

type Service struct {
	name   string
	reader *kafkago.Reader
	writer *kafkago.Writer
	hdl    Handler
}

func NewKafkaService(name, broker, requestTopic, responseTopic, groupID string, h Handler) *Service {
	dialer := kafkautil.NewDialer()
	transport := kafkautil.NewTransport(dialer)

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{broker},
		Topic:   requestTopic,
		GroupID: groupID,
		Dialer:  dialer,
		Logger: kafkago.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("[%s/reader] "+msg, append([]interface{}{name}, args...)...)
		}),
	})

	writer := &kafkago.Writer{
		Addr:      kafkago.TCP(broker),
		Topic:     responseTopic,
		Balancer:  &kafkago.LeastBytes{},
		Transport: transport,
		Logger: kafkago.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("[%s/writer] "+msg, append([]interface{}{name}, args...)...)
		}),
	}

	return &Service{name: name, reader: reader, writer: writer, hdl: h}
}

func (s *Service) Run(ctx context.Context) error {
	log.Printf("[%s] component service listening topic=%s", s.name, s.reader.Config().Topic)
	defer s.reader.Close()
	defer s.writer.Close()

	for {
		msg, err := s.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("[%s] read error: %v", s.name, err)
			continue
		}
		s.processMessage(ctx, msg)
	}
}

func (s *Service) processMessage(ctx context.Context, msg kafkago.Message) {
	var req models.Request
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		log.Printf("[%s] invalid request: %v", s.name, err)
		return
	}
	if req.CorrelationID == "" {
		req.CorrelationID = string(msg.Key)
	}

	resp, key := handleRequest(s.name, req, string(msg.Key), s.hdl)

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("[%s] cannot marshal response: %v", s.name, err)
		return
	}
	if err := s.writer.WriteMessages(ctx, kafkago.Message{Key: []byte(key), Value: data}); err != nil {
		log.Printf("[%s] response write error: %v", s.name, err)
		return
	}
	log.Printf("[%s] response sent correlation_id=%s success=%v", s.name, key, resp.Success)
}

func handleRequest(name string, req models.Request, fallbackKey string, h Handler) (models.Response, string) {
	resp, ok := h.Handle(req)
	if !ok {
		resp = models.Response{
			Action:        models.ResponseAction,
			Sender:        models.DefaultSender,
			CorrelationID: req.GetCorrelationID(),
			Success:       false,
			Error:         fmt.Sprintf("%s cannot handle action=%s", name, req.Action),
			Timestamp:     req.Timestamp,
		}
	}

	key := resp.CorrelationID
	if key == "" {
		key = req.GetCorrelationID()
	}
	if key == "" {
		key = fallbackKey
	}
	return resp, key
}
