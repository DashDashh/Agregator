package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OrderStatus is the order lifecycle status.
type OrderStatus string

const (
	StatusPending          OrderStatus = "pending"
	StatusSearching        OrderStatus = "searching"
	StatusMatched          OrderStatus = "matched"
	StatusConfirmed        OrderStatus = "confirmed"
	StatusCompletedPending OrderStatus = "completed_pending_confirmation"
	StatusCompleted        OrderStatus = "completed"
	StatusDispute          OrderStatus = "dispute"
)

type Order struct {
	ID               string      `json:"id"`
	CustomerID       string      `json:"customer_id"`
	Description      string      `json:"description"`
	Budget           float64     `json:"budget"`
	FromLat          float64     `json:"from_lat"`
	FromLon          float64     `json:"from_lon"`
	ToLat            float64     `json:"to_lat"`
	ToLon            float64     `json:"to_lon"`
	Status           OrderStatus `json:"status"`
	OperatorID       string      `json:"operator_id,omitempty"`
	OfferedPrice     float64     `json:"offered_price,omitempty"`
	MissionType      string      `json:"mission_type"`
	SecurityGoals    []string    `json:"security_goals,omitempty"`
	TopLeftLat       float64     `json:"top_left_lat,omitempty"`
	TopLeftLon       float64     `json:"top_left_lon,omitempty"`
	BottomRightLat   float64     `json:"bottom_right_lat,omitempty"`
	BottomRightLon   float64     `json:"bottom_right_lon,omitempty"`
	CommissionAmount float64     `json:"commission_amount,omitempty"`
	OperatorAmount   float64     `json:"operator_amount,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

type Operator struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	License      string `json:"license"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type Customer struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	PasswordHash string `json:"-"`
}

// Store wraps PostgreSQL access.
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
