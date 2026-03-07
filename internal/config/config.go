package config

import (
	"os"
)

type Config struct {
	KafkaBroker           string
	RequestTopic          string
	ResponseTopic         string
	ConsumerGroup         string
	DeadLetterTopic       string
	OperatorTopic         string // топик куда агрегатор пишет задания для эксплуатантов
	OperatorResponseTopic string // топик откуда агрегатор читает ответы эксплуатантов
	DatabaseURL           string
	MigrationsPath        string // путь к SQL-файлу миграции
}

func Load() *Config {
	return &Config{
		KafkaBroker:           getEnv("KAFKA_BROKER", "localhost:9092"),
		RequestTopic:          getEnv("KAFKA_REQUEST_TOPIC", "aggregator.requests"),
		ResponseTopic:         getEnv("KAFKA_RESPONSE_TOPIC", "aggregator.responses"),
		ConsumerGroup:         getEnv("KAFKA_CONSUMER_GROUP", "aggregator-group"),
		DeadLetterTopic:       getEnv("KAFKA_DLT_TOPIC", "aggregator.dead-letter"),
		OperatorTopic:         getEnv("KAFKA_OPERATOR_TOPIC", "operator.requests"),
		OperatorResponseTopic: getEnv("KAFKA_OPERATOR_RESPONSE_TOPIC", "operator.responses"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://aggregator:secret@localhost:5432/aggregator?sslmode=disable"),
		MigrationsPath:        getEnv("MIGRATIONS_PATH", "migrations/001_init.sql"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
