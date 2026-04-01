package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	OperatorTransport     string // kafka | both (MQTT только для operator.* топиков)

	MQTTBroker            string
	MQTTClientID          string
	MQTTUsername          string
	MQTTPassword          string
	MQTTOperatorTopic     string
	MQTTOperatorRespTopic string
	MQTTQoS               byte
}

func Load() *Config {
	const (
		defaultProtocolVersion   = "v1"
		defaultNamespace         = "Agregator"
		defaultServiceName       = "agregator_service"
		defaultClientName        = "aggregator_insurer"
		defaultInstanceID        = "local"
		defaultCommissionRate    = 0.1
		defaultMQTTQoS           = 1
		defaultOperatorTransport = "kafka"
	)

	protocolVersion := getEnv("KAFKA_PROTOCOL_VERSION", defaultProtocolVersion)
	namespace := getEnv("KAFKA_NAMESPACE", getEnv("KAFKA_SYSTEM_NAME", defaultNamespace))
	serviceName := getEnv("KAFKA_SERVICE_NAME", defaultServiceName)
	instanceID := getEnv("KAFKA_INSTANCE_ID", defaultInstanceID)

	topicPrefix := fmt.Sprintf("%s.%s.%s.%s", protocolVersion, namespace, instanceID, serviceName)
	defaultRequestTopic := topicPrefix + ".requests"
	defaultResponseTopic := topicPrefix + ".responses"
	defaultDLTTopic := topicPrefix + ".dead_letter"
	defaultOperatorTopic := topicPrefix + ".operator.requests"
	defaultOperatorResponseTopic := topicPrefix + ".operator.responses"
	defaultConsumerGroup := fmt.Sprintf("%s-%s-%s-%s-group", strings.ToLower(namespace), strings.ToLower(serviceName), instanceID, protocolVersion)

	defaultMQTTOperatorTopic := topicPrefix + ".operator.requests"
	defaultMQTTOperatorRespTopic := topicPrefix + ".operator.responses"

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
		OperatorTransport:     normalizeOperatorTransport(getEnv("OPERATOR_TRANSPORT", defaultOperatorTransport)),

		MQTTBroker:            getEnv("MQTT_BROKER", "mqtt:1883"),
		MQTTClientID:          getEnv("MQTT_CLIENT_ID", fmt.Sprintf("%s-%s-%s", defaultClientName, instanceID, "mqtt")),
		MQTTUsername:          getEnv("MQTT_USERNAME", ""),
		MQTTPassword:          getEnv("MQTT_PASSWORD", ""),
		MQTTOperatorTopic:     getEnv("MQTT_OPERATOR_TOPIC", defaultMQTTOperatorTopic),
		MQTTOperatorRespTopic: getEnv("MQTT_OPERATOR_RESPONSE_TOPIC", defaultMQTTOperatorRespTopic),
		MQTTQoS:               byte(getEnvFloat("MQTT_QOS", defaultMQTTQoS)),
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

func normalizeOperatorTransport(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "", "kafka":
		return "kafka"
	case "both", "kafka+mqtt", "mqtt+kafka", "kafka_mqtt", "kafka-mqtt":
		return "both"
	default:
		return v
	}
}

func (c *Config) Validate() error {
	switch c.OperatorTransport {
	case "kafka", "both":
		return nil
	default:
		return fmt.Errorf("unsupported OPERATOR_TRANSPORT=%q, allowed values: kafka, both", c.OperatorTransport)
	}
}

func (c *Config) UseMQTTForOperators() bool {
	return c.OperatorTransport == "both"
}
