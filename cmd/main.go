package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirilltahmazidi/aggregator/internal/config"
	"github.com/kirilltahmazidi/aggregator/internal/handler"
	"github.com/kirilltahmazidi/aggregator/internal/kafka"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[main] aggregator service starting")

	cfg := config.Load()
	log.Printf("[main] config: broker=%s request_topic=%s response_topic=%s",
		cfg.KafkaBroker, cfg.RequestTopic, cfg.ResponseTopic)

	h := handler.New()
	svc := kafka.NewService(cfg, h)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := svc.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("[main] service exited with error: %v", err)
	}

	log.Println("[main] aggregator service stopped gracefully")
}
