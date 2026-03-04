package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OrderStatus — статус заказа на каждом этапе жизненного цикла
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"   // ждёт исполнителя
	StatusSearching OrderStatus = "searching" // идёт поиск эксплуатанта
	StatusMatched   OrderStatus = "matched"   // исполнитель найден
	StatusConfirmed OrderStatus = "confirmed" // контракт подписан
	StatusCompleted OrderStatus = "completed" // заказ выполнен
	StatusDispute   OrderStatus = "dispute"   // открыт спор
)

// Order — запись о заказе
type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	Description string      `json:"description"`
	Budget      float64     `json:"budget"`
	FromLat     float64     `json:"from_lat"`
	FromLon     float64     `json:"from_lon"`
	ToLat       float64     `json:"to_lat"`
	ToLon       float64     `json:"to_lon"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
}

// Operator — зарегистрированный эксплуатант
type Operator struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	License string `json:"license"`
	Email   string `json:"email"`
}

// Customer — зарегистрированный заказчик
type Customer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// Store — хранилище на основе PostgreSQL
type Store struct {
	db *sql.DB // пул соединений к базе данных
}

// New открывает соединение с базой данных и проверяет что оно живое.
func New(databaseURL string) (*Store, error) {
	// просто создаёт пул
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// Ping — вот здесь происходит первое подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return &Store{db: db}, nil
}

// Close закрывает пул соединений при остановке сервиса
func (s *Store) Close() error {
	return s.db.Close()
}

// RunMigrations выполняет SQL-схему — запускаем при старте сервиса.
func (s *Store) RunMigrations(sqlText string) error {
	_, err := s.db.Exec(sqlText)
	return err
}

// Orders
func (s *Store) SaveOrder(o *Order) error {
	// ON CONFLICT (id) DO UPDATE — upsert: вставить или обновить если id уже есть
	_, err := s.db.Exec(`
		INSERT INTO orders (id, customer_id, description, budget, from_lat, from_lon, to_lat, to_lon, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status
	`, o.ID, o.CustomerID, o.Description, o.Budget,
		o.FromLat, o.FromLon, o.ToLat, o.ToLon,
		string(o.Status), o.CreatedAt)
	return err
}

func (s *Store) GetOrder(id string) (*Order, bool) {
	o := &Order{}
	err := s.db.QueryRow(`
		SELECT id, customer_id, description, budget, from_lat, from_lon, to_lat, to_lon, status, created_at
		FROM orders WHERE id = $1
	`, id).Scan(
		&o.ID, &o.CustomerID, &o.Description, &o.Budget,
		&o.FromLat, &o.FromLon, &o.ToLat, &o.ToLon,
		&o.Status, &o.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, false // не нашли — возвращаем false как раньше
	}
	if err != nil {
		return nil, false
	}
	return o, true
}

func (s *Store) ListOrders() []*Order {
	rows, err := s.db.Query(`
		SELECT id, customer_id, description, budget, from_lat, from_lon, to_lat, to_lon, status, created_at
		FROM orders ORDER BY created_at DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() { // идём по строкам результата одна за одной
		o := &Order{}
		if err := rows.Scan(
			&o.ID, &o.CustomerID, &o.Description, &o.Budget,
			&o.FromLat, &o.FromLon, &o.ToLat, &o.ToLon,
			&o.Status, &o.CreatedAt,
		); err != nil {
			continue
		}
		orders = append(orders, o)
	}
	return orders
}

func (s *Store) UpdateOrderStatus(id string, status OrderStatus) bool {
	res, err := s.db.Exec(`UPDATE orders SET status = $1 WHERE id = $2`, string(status), id)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected() // сколько строк было изменено
	return n > 0               // если 0 — такого заказа нет
}

// Operators

func (s *Store) SaveOperator(op *Operator) error {
	_, err := s.db.Exec(`
		INSERT INTO operators (id, name, license, email)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, op.ID, op.Name, op.License, op.Email)
	return err
}

func (s *Store) GetOperator(id string) (*Operator, bool) {
	op := &Operator{}
	err := s.db.QueryRow(`
		SELECT id, name, license, email FROM operators WHERE id = $1
	`, id).Scan(&op.ID, &op.Name, &op.License, &op.Email)
	if err != nil {
		return nil, false
	}
	return op, true
}

// Customers

func (s *Store) SaveCustomer(c *Customer) error {
	_, err := s.db.Exec(`
		INSERT INTO customers (id, name, email, phone)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, c.ID, c.Name, c.Email, c.Phone)
	return err
}

func (s *Store) GetCustomer(id string) (*Customer, bool) {
	c := &Customer{}
	err := s.db.QueryRow(`
		SELECT id, name, email, phone FROM customers WHERE id = $1
	`, id).Scan(&c.ID, &c.Name, &c.Email, &c.Phone)
	if err != nil {
		return nil, false
	}
	return c, true
}
