package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kirilltahmazidi/aggregator/internal/api"
	"github.com/kirilltahmazidi/aggregator/internal/config"
	"github.com/kirilltahmazidi/aggregator/internal/handler"
	"github.com/kirilltahmazidi/aggregator/internal/kafka"
	"github.com/kirilltahmazidi/aggregator/internal/store"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[main] aggregator service starting")

	cfg := config.Load()
	log.Printf("[main] config: broker=%s request_topic=%s response_topic=%s",
		cfg.KafkaBroker, cfg.RequestTopic, cfg.ResponseTopic)

	// Подключаемся к PostgreSQL
	s, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[main] cannot connect to database: %v", err)
	}
	defer s.Close()
	log.Println("[main] connected to database")

	// Запускаем миграции — создаём таблицы если их нет
	migration, err := os.ReadFile(cfg.MigrationsPath)
	if err != nil {
		log.Fatalf("[main] cannot read migration file: %v", err)
	}
	if err := s.RunMigrations(string(migration)); err != nil {
		log.Fatalf("[main] migration failed: %v", err)
	}
	log.Println("[main] migrations applied")

	// Kafka-сервис — передаём store чтобы он мог обновлять статусы заказов
	h := handler.New()
	svc := kafka.NewService(cfg, h, s)

	// HTTP-сервер для фронтенда — передаём kafka сервис как Publisher
	apiHandler := api.NewHandler(s, svc)
	router := api.NewRouter(apiHandler)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Запускаем HTTP-сервер в горутине — параллельно с Kafka
	go func() {
		log.Println("[main] HTTP server listening on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[main] HTTP server error: %v", err)
		}
	}()

	// Когда получим сигнал остановки — корректно останавливаем HTTP
	go func() {
		<-ctx.Done()
		log.Println("[main] shutting down HTTP server...")
		httpServer.Shutdown(context.Background()) //nolint:errcheck
	}()

	// Kafka-цикл блокирует до остановки
	if err := svc.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("[main] service exited with error: %v", err)
	}

	log.Println("[main] aggregator service stopped gracefully")
}
