package store

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func newMockStore(t *testing.T) (*Store, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return &Store{db: db}, mock, func() { db.Close() }
}

func TestCustomerStoreMethods(t *testing.T) {
	st, mock, closeDB := newMockStore(t)
	defer closeDB()

	customer := &Customer{
		ID:           "customer-1",
		Name:         "Ivan",
		Email:        "ivan@example.com",
		Phone:        "+79001234567",
		PasswordHash: "hash",
	}
	mock.ExpectExec("INSERT INTO customers").
		WithArgs(customer.ID, customer.Name, customer.Email, customer.Phone, customer.PasswordHash).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := st.SaveCustomer(customer); err != nil {
		t.Fatalf("SaveCustomer returned error: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, phone, password_hash FROM customers WHERE id = $1")).
		WithArgs("customer-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "password_hash"}).
			AddRow(customer.ID, customer.Name, customer.Email, customer.Phone, customer.PasswordHash))
	got, ok := st.GetCustomer("customer-1")
	if !ok || got.ID != "customer-1" {
		t.Fatalf("GetCustomer = %+v/%v", got, ok)
	}

	mock.ExpectQuery("SELECT id, name, email, phone, password_hash FROM customers").
		WithArgs("ivan@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "password_hash"}).
			AddRow(customer.ID, customer.Name, customer.Email, customer.Phone, customer.PasswordHash))
	got, ok = st.GetCustomerByEmail("ivan@example.com")
	if !ok || got.Email != "ivan@example.com" {
		t.Fatalf("GetCustomerByEmail = %+v/%v", got, ok)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE customers SET password_hash = $1 WHERE id = $2")).
		WithArgs("new-hash", "customer-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !st.SetCustomerPasswordHash("customer-1", "new-hash") {
		t.Fatal("SetCustomerPasswordHash returned false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestOperatorStoreMethods(t *testing.T) {
	st, mock, closeDB := newMockStore(t)
	defer closeDB()

	operator := &Operator{
		ID:           "operator-1",
		Name:         "Operator",
		License:      "LIC-1",
		Email:        "op@example.com",
		PasswordHash: "hash",
	}
	mock.ExpectExec("INSERT INTO operators").
		WithArgs(operator.ID, operator.Name, operator.License, operator.Email, operator.PasswordHash).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := st.SaveOperator(operator); err != nil {
		t.Fatalf("SaveOperator returned error: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, license, email, password_hash FROM operators WHERE id = $1")).
		WithArgs("operator-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "license", "email", "password_hash"}).
			AddRow(operator.ID, operator.Name, operator.License, operator.Email, operator.PasswordHash))
	got, ok := st.GetOperator("operator-1")
	if !ok || got.ID != "operator-1" {
		t.Fatalf("GetOperator = %+v/%v", got, ok)
	}

	mock.ExpectQuery("SELECT id, name, license, email, password_hash FROM operators").
		WithArgs("op@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "license", "email", "password_hash"}).
			AddRow(operator.ID, operator.Name, operator.License, operator.Email, operator.PasswordHash))
	got, ok = st.GetOperatorByEmail("op@example.com")
	if !ok || got.Email != "op@example.com" {
		t.Fatalf("GetOperatorByEmail = %+v/%v", got, ok)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE operators SET password_hash = $1 WHERE id = $2")).
		WithArgs("new-hash", "operator-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !st.SetOperatorPasswordHash("operator-1", "new-hash") {
		t.Fatal("SetOperatorPasswordHash returned false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestOrderStoreTransitions(t *testing.T) {
	st, mock, closeDB := newMockStore(t)
	defer closeDB()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE orders SET status = $1 WHERE id = $2")).
		WithArgs(string(StatusSearching), "order-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !st.UpdateOrderStatus("order-1", StatusSearching) {
		t.Fatal("UpdateOrderStatus returned false")
	}

	mock.ExpectExec("UPDATE orders").
		WithArgs(string(StatusConfirmed), float64(1000), float64(100), float64(900), "order-1", string(StatusMatched), "operator-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !st.ConfirmPrice("order-1", "operator-1", 1000, 100) {
		t.Fatal("ConfirmPrice returned false")
	}

	if st.ConfirmPrice("order-1", "operator-1", 0, 100) {
		t.Fatal("ConfirmPrice accepted zero price")
	}

	mock.ExpectExec("UPDATE orders").
		WithArgs(string(StatusCompleted), "order-1", string(StatusCompletedPending)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !st.ConfirmCompletion("order-1") {
		t.Fatal("ConfirmCompletion returned false")
	}

	mock.ExpectExec("UPDATE orders").
		WithArgs(string(StatusMatched), "operator-1", float64(1000), "order-1", string(StatusSearching), string(StatusPending)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !st.SetOperatorOffer("order-1", "operator-1", 1000) {
		t.Fatal("SetOperatorOffer returned false")
	}

	mock.ExpectExec("UPDATE orders").
		WithArgs(string(StatusDispute), "order-1", string(StatusConfirmed)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if !st.ProcessOrderResult("order-1", false) {
		t.Fatal("ProcessOrderResult returned false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestOrderStoreReadMethods(t *testing.T) {
	st, mock, closeDB := newMockStore(t)
	defer closeDB()

	createdAt := time.Now().UTC()
	rows := orderRows().AddRow(
		"order-1", "customer-1", "docs", float64(1000),
		float64(1), float64(2), float64(3), float64(4),
		string(StatusSearching), "operator-1", float64(900),
		"delivery", pq.StringArray{"CB1"},
		float64(10), float64(20), float64(30), float64(40),
		float64(90), float64(810), createdAt,
	)
	mock.ExpectQuery("SELECT id, customer_id, description, budget").
		WithArgs("order-1").
		WillReturnRows(rows)
	order, ok := st.GetOrder("order-1")
	if !ok || order.ID != "order-1" || len(order.SecurityGoals) != 1 {
		t.Fatalf("GetOrder = %+v/%v", order, ok)
	}

	mock.ExpectQuery("SELECT id, customer_id, description, budget").
		WillReturnRows(orderRows().AddRow(
			"order-2", "customer-1", "docs", float64(1000),
			float64(1), float64(2), float64(3), float64(4),
			string(StatusPending), "", float64(0),
			"delivery", pq.StringArray{},
			float64(0), float64(0), float64(0), float64(0),
			float64(0), float64(0), createdAt,
		))
	if got := st.ListOrders(); len(got) != 1 || got[0].ID != "order-2" {
		t.Fatalf("ListOrders = %+v", got)
	}

	mock.ExpectQuery("SELECT id, customer_id, description, budget").
		WithArgs("customer-1").
		WillReturnRows(orderRows().AddRow(
			"order-3", "customer-1", "docs", float64(1000),
			float64(1), float64(2), float64(3), float64(4),
			string(StatusPending), "", float64(0),
			"delivery", pq.StringArray{},
			float64(0), float64(0), float64(0), float64(0),
			float64(0), float64(0), createdAt,
		))
	if got := st.ListOrdersByCustomer("customer-1"); len(got) != 1 || got[0].ID != "order-3" {
		t.Fatalf("ListOrdersByCustomer = %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestIncidentStoreMethods(t *testing.T) {
	st, mock, closeDB := newMockStore(t)
	defer closeDB()

	createdAt := time.Now().UTC()
	incident := &Incident{
		ID:           "incident-1",
		OrderID:      "order-1",
		OperatorID:   "operator-1",
		ReporterID:   "customer-1",
		Reason:       "drone_lost",
		Description:  "failed",
		DamageAmount: 5000,
		Status:       "registered",
		CreatedAt:    createdAt,
	}
	mock.ExpectExec("INSERT INTO incidents").
		WithArgs(incident.ID, incident.OrderID, incident.OperatorID, incident.ReporterID, incident.Reason, incident.Description, incident.DamageAmount, incident.Status, incident.CreatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE orders SET status = $1 WHERE id = $2")).
		WithArgs(string(StatusDispute), "order-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := st.RegisterIncident(incident); err != nil {
		t.Fatalf("RegisterIncident returned error: %v", err)
	}

	mock.ExpectQuery("SELECT id, order_id, operator_id, reporter_id").
		WithArgs("order-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "operator_id", "reporter_id", "reason", "description", "damage_amount", "status", "created_at"}).
			AddRow(incident.ID, incident.OrderID, incident.OperatorID, incident.ReporterID, incident.Reason, incident.Description, incident.DamageAmount, incident.Status, incident.CreatedAt))
	got := st.ListIncidentsByOrder("order-1")
	if len(got) != 1 || got[0].ID != "incident-1" {
		t.Fatalf("ListIncidentsByOrder = %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func orderRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "customer_id", "description", "budget",
		"from_lat", "from_lon", "to_lat", "to_lon",
		"status", "operator_id", "offered_price",
		"mission_type", "security_goals",
		"top_left_lat", "top_left_lon", "bottom_right_lat", "bottom_right_lon",
		"commission_amount", "operator_amount", "created_at",
	})
}
