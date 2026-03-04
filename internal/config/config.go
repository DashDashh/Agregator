package config

import (
	"os"
)

type Config struct {
	KafkaBroker     string
	RequestTopic    string
	ResponseTopic   string
	ConsumerGroup   string
	DeadLetterTopic string
	OperatorTopic   string // топик куда агрегатор пишет задания для эксплуатантов
}

func Load() *Config {
	return &Config{
		KafkaBroker:     getEnv("KAFKA_BROKER", "localhost:9092"),
		RequestTopic:    getEnv("KAFKA_REQUEST_TOPIC", "aggregator.requests"),
		ResponseTopic:   getEnv("KAFKA_RESPONSE_TOPIC", "aggregator.responses"),
		ConsumerGroup:   getEnv("KAFKA_CONSUMER_GROUP", "aggregator-group"),
		DeadLetterTopic: getEnv("KAFKA_DLT_TOPIC", "aggregator.dead-letter"),
		OperatorTopic:   getEnv("KAFKA_OPERATOR_TOPIC", "operator.requests"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
