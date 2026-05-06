package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

type fakeRegistryStore struct {
	customer *store.Customer
	operator *store.Operator
}

func (f *fakeRegistryStore) SaveCustomer(c *store.Customer) error {
	f.customer = c
	return nil
}

func (f *fakeRegistryStore) GetCustomer(id string) (*store.Customer, bool) {
	if f.customer != nil && f.customer.ID == id {
		return f.customer, true
	}
	return nil, false
}

func (f *fakeRegistryStore) GetCustomerByEmail(email string) (*store.Customer, bool) {
	if f.customer != nil && f.customer.Email == email {
		return f.customer, true
	}
	return nil, false
}

func (f *fakeRegistryStore) SetCustomerPasswordHash(id, passwordHash string) bool {
	if f.customer != nil && f.customer.ID == id {
		f.customer.PasswordHash = passwordHash
		return true
	}
	return false
}

func (f *fakeRegistryStore) SaveOperator(op *store.Operator) error {
	f.operator = op
	return nil
}

func (f *fakeRegistryStore) GetOperator(id string) (*store.Operator, bool) {
	if f.operator != nil && f.operator.ID == id {
		return f.operator, true
	}
	return nil, false
}

func (f *fakeRegistryStore) GetOperatorByEmail(email string) (*store.Operator, bool) {
	if f.operator != nil && f.operator.Email == email {
		return f.operator, true
	}
	return nil, false
}

func (f *fakeRegistryStore) SetOperatorPasswordHash(id, passwordHash string) bool {
	if f.operator != nil && f.operator.ID == id {
		f.operator.PasswordHash = passwordHash
		return true
	}
	return false
}

func TestLoginCustomerSuccess(t *testing.T) {
	hash, err := auth.HashPassword("strongpass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{customer: &store.Customer{
		ID:           "customer-1",
		Name:         "Ivan",
		Email:        "ivan@example.com",
		PasswordHash: hash,
	}}, "test-secret")

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{
		"role": "customer",
		"email": "ivan@example.com",
		"password": "strongpass123"
	}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("body does not contain token: %s", rec.Body.String())
	}
}

func TestLoginRejectsWrongRole(t *testing.T) {
	h := NewHandler(&fakeRegistryStore{}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"role":"admin"}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
