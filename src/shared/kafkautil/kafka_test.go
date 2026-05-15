package kafkautil

import (
	"crypto/tls"
	"testing"
)

func TestNewDialerWithoutCredentials(t *testing.T) {
	t.Setenv("BROKER_USER", "")
	t.Setenv("BROKER_PASSWORD", "")
	t.Setenv("KAFKA_TLS_ENABLED", "")

	dialer := NewDialer()
	if dialer == nil {
		t.Fatal("NewDialer returned nil")
	}
	if dialer.SASLMechanism != nil {
		t.Fatal("SASLMechanism is set without credentials")
	}
	if dialer.TLS != nil {
		t.Fatal("TLS is set when KAFKA_TLS_ENABLED is not true")
	}
}

func TestNewDialerWithCredentialsAndTLS(t *testing.T) {
	t.Setenv("BROKER_USER", "user")
	t.Setenv("BROKER_PASSWORD", "pass")
	t.Setenv("KAFKA_TLS_ENABLED", "true")

	dialer := NewDialer()
	if dialer.SASLMechanism == nil {
		t.Fatal("SASLMechanism is nil with credentials")
	}
	if dialer.TLS == nil {
		t.Fatal("TLS is nil with KAFKA_TLS_ENABLED=true")
	}
	if dialer.TLS.MinVersion != tls.VersionTLS12 {
		t.Fatalf("TLS MinVersion = %v, want TLS 1.2", dialer.TLS.MinVersion)
	}
}

func TestNewTransportCopiesDialerSecuritySettings(t *testing.T) {
	dialer := NewDialer()
	dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}

	transport := NewTransport(dialer)
	if transport.TLS != dialer.TLS {
		t.Fatal("transport TLS does not point to dialer TLS")
	}

	empty := NewTransport(nil)
	if empty == nil {
		t.Fatal("NewTransport(nil) returned nil")
	}
}
