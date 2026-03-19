package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	KafkaBroker           string
	RequestTopic          string
	ResponseTopic         string
	ConsumerGroup         string
	DeadLetterTopic       string
	OperatorTopic         string // топик куда агрегатор пишет задания для эксплуатантов
	OperatorResponseTopic string // топик откуда агрегатор читает ответы эксплуатантов
	CommissionRate        float64
	DatabaseURL           string
	MigrationsPath        string // путь к SQL-файлу миграции
}

func Load() *Config {
	const (
		defaultProtocolVersion = "v1"
		defaultSystemName      = "aggregator_insurer"
		defaultInstanceID      = "local"
		defaultCommissionRate  = 0.1
	)

	protocolVersion := getEnv("KAFKA_PROTOCOL_VERSION", defaultProtocolVersion)
	systemName := getEnv("KAFKA_SYSTEM_NAME", defaultSystemName)
	instanceID := getEnv("KAFKA_INSTANCE_ID", defaultInstanceID)

	topicPrefix := fmt.Sprintf("%s.%s.%s", protocolVersion, systemName, instanceID)
	defaultRequestTopic := topicPrefix + ".aggregator.requests"
	defaultResponseTopic := topicPrefix + ".aggregator.responses"
	defaultDLTTopic := topicPrefix + ".aggregator.dead_letter"
	defaultOperatorTopic := topicPrefix + ".operator.requests"
	defaultOperatorResponseTopic := topicPrefix + ".operator.responses"
	defaultConsumerGroup := fmt.Sprintf("%s-%s-%s-group", systemName, instanceID, protocolVersion)

	commissionRate := getEnvFloat("COMMISSION_RATE", defaultCommissionRate)

	return &Config{
		KafkaBroker:           getEnv("KAFKA_BROKER", "localhost:9092"),
		RequestTopic:          getEnv("KAFKA_REQUEST_TOPIC", defaultRequestTopic),
		ResponseTopic:         getEnv("KAFKA_RESPONSE_TOPIC", defaultResponseTopic),
		ConsumerGroup:         getEnv("KAFKA_CONSUMER_GROUP", defaultConsumerGroup),
		DeadLetterTopic:       getEnv("KAFKA_DLT_TOPIC", defaultDLTTopic),
		OperatorTopic:         getEnv("KAFKA_OPERATOR_TOPIC", defaultOperatorTopic),
		OperatorResponseTopic: getEnv("KAFKA_OPERATOR_RESPONSE_TOPIC", defaultOperatorResponseTopic),
		CommissionRate:        commissionRate,
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

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
