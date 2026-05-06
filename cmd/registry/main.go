package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirilltahmazidi/aggregator/src/gateway/config"
	"github.com/kirilltahmazidi/aggregator/src/registry_component"
	"github.com/kirilltahmazidi/aggregator/src/shared/componentbus"
)

func main() {
	run("registry", registry_component.Topic, registry_component.NewHandler())
}

func run(name, topic string, handler componentbus.Handler) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[%s] invalid config: %v", name, err)
	}

	groupID := cfg.ConsumerGroup + "-" + name
	service := componentbus.NewKafkaService(name, cfg.KafkaBroker, topic, cfg.ResponseTopic, groupID, handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := service.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("[%s] service stopped: %v", name, err)
	}
}
