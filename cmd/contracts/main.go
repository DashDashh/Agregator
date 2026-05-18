package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirilltahmazidi/aggregator/src/contracts_component"
	"github.com/kirilltahmazidi/aggregator/src/gateway/config"
	"github.com/kirilltahmazidi/aggregator/src/shared/componentbus"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

func main() {
	name := "contracts"
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[%s] invalid config: %v", name, err)
	}
	st, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[%s] cannot connect to database: %v", name, err)
	}
	defer st.Close()
	runMigrations(name, cfg, st)

	run(name, contracts_component.Topic, contracts_component.NewStoreHandler(st, cfg.CommissionRate), cfg)
}

func run(name, topic string, handler componentbus.Handler, cfg *config.Config) {
	groupID := cfg.ConsumerGroup + "-" + name
	service := componentbus.NewKafkaService(name, cfg.KafkaBroker, topic, cfg.ResponseTopic, groupID, handler)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := service.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("[%s] service stopped: %v", name, err)
	}
}

func runMigrations(name string, cfg *config.Config, st *store.Store) {
	migration, err := os.ReadFile(cfg.MigrationsPath)
	if err != nil {
		log.Fatalf("[%s] cannot read migration file: %v", name, err)
	}
	if err := st.RunMigrations(string(migration)); err != nil {
		log.Fatalf("[%s] migration failed: %v", name, err)
	}
}
