package config

import "testing"

func TestLoadAppliesNamespaceDefaults(t *testing.T) {
	t.Setenv("SYSTEM_NAMESPACE", ".team.alpha.")
	t.Setenv("OPERATOR_TRANSPORT", "")
	t.Setenv("KAFKA_REQUEST_TOPIC", "")
	t.Setenv("KAFKA_RESPONSE_TOPIC", "")
	t.Setenv("KAFKA_OPERATOR_TOPIC", "")
	t.Setenv("KAFKA_OPERATOR_RESPONSE_TOPIC", "")
	t.Setenv("KAFKA_CONSUMER_GROUP", "")
	t.Setenv("MQTT_CLIENT_ID", "")

	cfg := Load()

	if cfg.RequestTopic != "team.alpha.systems.agregator" {
		t.Fatalf("RequestTopic = %q", cfg.RequestTopic)
	}
	if cfg.ResponseTopic != "team.alpha.components.agregator.responses" {
		t.Fatalf("ResponseTopic = %q", cfg.ResponseTopic)
	}
	if cfg.OperatorTopic != "team.alpha.components.agregator.operator.requests" {
		t.Fatalf("OperatorTopic = %q", cfg.OperatorTopic)
	}
	if cfg.OperatorResponseTopic != "team.alpha.components.agregator.operator.responses" {
		t.Fatalf("OperatorResponseTopic = %q", cfg.OperatorResponseTopic)
	}
	if cfg.ConsumerGroup != "team-alpha-agregator-group" {
		t.Fatalf("ConsumerGroup = %q", cfg.ConsumerGroup)
	}
	if cfg.MQTTClientID != "agregator-team.alpha-mqtt" {
		t.Fatalf("MQTTClientID = %q", cfg.MQTTClientID)
	}
}

func TestLoadNormalizesOperatorTransportAliases(t *testing.T) {
	t.Setenv("OPERATOR_TRANSPORT", " kafka+mqtt ")

	cfg := Load()

	if cfg.OperatorTransport != "both" {
		t.Fatalf("OperatorTransport = %q, want both", cfg.OperatorTransport)
	}
	if !cfg.UseMQTTForOperators() {
		t.Fatal("UseMQTTForOperators returned false for both transport")
	}
}

func TestValidateRejectsUnsupportedTransport(t *testing.T) {
	cfg := &Config{OperatorTransport: "mqtt-only"}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate accepted unsupported transport")
	}
}

func TestLoadUsesFallbackForInvalidFloatEnv(t *testing.T) {
	t.Setenv("COMMISSION_RATE", "not-a-number")
	t.Setenv("MQTT_QOS", "not-a-number")

	cfg := Load()

	if cfg.CommissionRate != 0.1 {
		t.Fatalf("CommissionRate = %v, want fallback 0.1", cfg.CommissionRate)
	}
	if cfg.MQTTQoS != 1 {
		t.Fatalf("MQTTQoS = %v, want fallback 1", cfg.MQTTQoS)
	}
}
