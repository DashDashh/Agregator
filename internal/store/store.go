package store

import (
	"sync"
	"time"
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

// Order — запись о заказе в памяти
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

// Store — потокобезопасное хранилище в памяти (заглушка вместо БД)
type Store struct {
	mu        sync.RWMutex
	orders    map[string]*Order
	operators map[string]*Operator
	customers map[string]*Customer
}

func New() *Store {
	return &Store{
		orders:    make(map[string]*Order),
		operators: make(map[string]*Operator),
		customers: make(map[string]*Customer),
	}
}

// Orders
// Просто кладем заказы по ID
func (s *Store) SaveOrder(o *Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.ID] = o
}

// Получаем заказ по ID, возвращаем его и флаг наличия
func (s *Store) GetOrder(id string) (*Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	return o, ok
}

// Получить все заказы
func (s *Store) ListOrders() []*Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Order, 0, len(s.orders))
	for _, o := range s.orders {
		out = append(out, o)
	}
	return out
}

// Обновляем статус заказа
func (s *Store) UpdateOrderStatus(id string, status OrderStatus) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o, ok := s.orders[id]; ok {
		o.Status = status
		return true
	}
	return false
}

// Operators

func (s *Store) SaveOperator(op *Operator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.operators[op.ID] = op
}

func (s *Store) GetOperator(id string) (*Operator, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	op, ok := s.operators[id]
	return op, ok
}

// Customers

func (s *Store) SaveCustomer(c *Customer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customers[c.ID] = c
}

func (s *Store) GetCustomer(id string) (*Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.customers[id]
	return c, ok
}
