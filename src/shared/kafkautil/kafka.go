package kafkautil

import (
	"crypto/tls"
	"os"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func NewDialer() *kafkago.Dialer {
	dialer := &kafkago.Dialer{}

	username := os.Getenv("BROKER_USER")
	password := os.Getenv("BROKER_PASSWORD")
	if username == "" && password == "" {
		return dialer
	}

	dialer.SASLMechanism = plain.Mechanism{
		Username: username,
		Password: password,
	}

	if os.Getenv("KAFKA_TLS_ENABLED") == "true" {
		dialer.TLS = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return dialer
}

func NewTransport(dialer *kafkago.Dialer) *kafkago.Transport {
	transport := &kafkago.Transport{}
	if dialer == nil {
		return transport
	}

	transport.TLS = dialer.TLS
	transport.SASL = dialer.SASLMechanism
	return transport
}
