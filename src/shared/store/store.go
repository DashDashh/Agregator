package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kirilltahmazidi/aggregator/src/shared/domain"
)

const (
	StatusPending          = domain.StatusPending
	StatusSearching        = domain.StatusSearching
	StatusMatched          = domain.StatusMatched
	StatusConfirmed        = domain.StatusConfirmed
	StatusCompletedPending = domain.StatusCompletedPending
	StatusCompleted        = domain.StatusCompleted
	StatusDispute          = domain.StatusDispute
)

type OrderStatus = domain.OrderStatus
type Order = domain.Order
type Operator = domain.Operator
type Customer = domain.Customer

// Store инкапсулирует доступ к PostgreSQL.
type Store struct {
	db *sql.DB
}

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) RunMigrations(sqlText string) error {
	_, err := s.db.Exec(sqlText)
	return err
}
