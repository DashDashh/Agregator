package httpapi

import (
	"encoding/json"
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
	drone    *store.Drone
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

func (f *fakeRegistryStore) SaveDrone(drone *store.Drone) error {
	f.drone = drone
	return nil
}

func (f *fakeRegistryStore) ListDronesByOperator(operatorID string) []*store.Drone {
	if f.drone != nil && f.drone.OperatorID == operatorID {
		return []*store.Drone{f.drone}
	}
	return nil
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

func TestLoginOperatorSuccess(t *testing.T) {
	hash, err := auth.HashPassword("strongpass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{operator: &store.Operator{
		ID:           "operator-1",
		Name:         "Operator",
		Email:        "op@example.com",
		PasswordHash: hash,
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{
		"role": "operator",
		"email": "op@example.com",
		"password": "strongpass123"
	}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"role":"operator"`) {
		t.Fatalf("body does not contain operator role: %s", rec.Body.String())
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

func TestLoginRejectsWrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("strongpass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{customer: &store.Customer{
		ID:           "customer-1",
		Email:        "ivan@example.com",
		PasswordHash: hash,
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{
		"role": "customer",
		"email": "ivan@example.com",
		"password": "badpass"
	}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLoginRejectsUnknownUser(t *testing.T) {
	h := NewHandler(&fakeRegistryStore{}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{
		"role": "customer",
		"email": "missing@example.com",
		"password": "strongpass123"
	}`))
	rec := httptest.NewRecorder()

	h.Login(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRegisterCustomerCreatesUserAndToken(t *testing.T) {
	repo := &fakeRegistryStore{}
	h := NewHandler(repo, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{
		"name": "Ivan",
		"email": "ivan@example.com",
		"phone": "+79001234567",
		"password": "strongpass123"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterCustomer(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if repo.customer == nil {
		t.Fatal("customer was not saved")
	}
	if repo.customer.PasswordHash == "" || repo.customer.PasswordHash == "strongpass123" {
		t.Fatal("customer password was not hashed")
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("cannot decode response: %v", err)
	}
	if body["token"] == "" {
		t.Fatalf("token is empty in response: %s", rec.Body.String())
	}
	if body["role"] != "customer" {
		t.Fatalf("role = %v, want customer", body["role"])
	}
}

func TestRegisterCustomerRejectsMissingFields(t *testing.T) {
	h := NewHandler(&fakeRegistryStore{}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{"email":"ivan@example.com"}`))
	rec := httptest.NewRecorder()

	h.RegisterCustomer(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegisterCustomerReturnsExistingForSamePassword(t *testing.T) {
	hash, err := auth.HashPassword("strongpass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	repo := &fakeRegistryStore{customer: &store.Customer{
		ID:           "customer-1",
		Name:         "Ivan",
		Email:        "ivan@example.com",
		PasswordHash: hash,
	}}
	h := NewHandler(repo, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{
		"name": "Ivan",
		"email": "ivan@example.com",
		"password": "strongpass123"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterCustomer(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"existing":true`) {
		t.Fatalf("body does not mark existing user: %s", rec.Body.String())
	}
}

func TestRegisterCustomerRejectsExistingWithDifferentPassword(t *testing.T) {
	hash, err := auth.HashPassword("strongpass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{customer: &store.Customer{
		ID:           "customer-1",
		Email:        "ivan@example.com",
		PasswordHash: hash,
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{
		"name": "Ivan",
		"email": "ivan@example.com",
		"password": "differentpass"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterCustomer(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestRegisterOperatorCreatesUserAndToken(t *testing.T) {
	repo := &fakeRegistryStore{}
	h := NewHandler(repo, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/operators", strings.NewReader(`{
		"name": "Operator",
		"license": "LIC-1",
		"email": "op@example.com",
		"password": "strongpass123"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterOperator(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if repo.operator == nil {
		t.Fatal("operator was not saved")
	}
	if repo.operator.PasswordHash == "" || repo.operator.PasswordHash == "strongpass123" {
		t.Fatal("operator password was not hashed")
	}
	if !strings.Contains(rec.Body.String(), `"role":"operator"`) {
		t.Fatalf("body does not contain operator role: %s", rec.Body.String())
	}
}

func TestRegisterOperatorRejectsMissingFields(t *testing.T) {
	h := NewHandler(&fakeRegistryStore{}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/operators", strings.NewReader(`{"name":"Operator"}`))
	rec := httptest.NewRecorder()

	h.RegisterOperator(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegisterOperatorReturnsExistingForSamePassword(t *testing.T) {
	hash, err := auth.HashPassword("strongpass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{operator: &store.Operator{
		ID:           "operator-1",
		Name:         "Operator",
		License:      "LIC-1",
		Email:        "op@example.com",
		PasswordHash: hash,
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/operators", strings.NewReader(`{
		"name": "Operator",
		"license": "LIC-1",
		"email": "op@example.com",
		"password": "strongpass123"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterOperator(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"existing":true`) {
		t.Fatalf("body does not mark existing operator: %s", rec.Body.String())
	}
}

func TestRegisterOperatorRejectsExistingWithDifferentPassword(t *testing.T) {
	hash, err := auth.HashPassword("strongpass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{operator: &store.Operator{
		ID:           "operator-1",
		Email:        "op@example.com",
		PasswordHash: hash,
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodPost, "/operators", strings.NewReader(`{
		"name": "Operator",
		"license": "LIC-1",
		"email": "op@example.com",
		"password": "differentpass"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterOperator(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestGetCustomerAllowsOwnProfile(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{customer: &store.Customer{
		ID:    "customer-1",
		Name:  "Ivan",
		Email: "ivan@example.com",
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/customers/customer-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.GetCustomer(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestFindCustomerReturnsCustomer(t *testing.T) {
	h := NewHandler(&fakeRegistryStore{customer: &store.Customer{
		ID:    "customer-1",
		Email: "ivan@example.com",
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/customers?email=ivan@example.com", nil)
	rec := httptest.NewRecorder()

	h.FindCustomer(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestFindCustomerRejectsMissingEmail(t *testing.T) {
	h := NewHandler(&fakeRegistryStore{}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	rec := httptest.NewRecorder()

	h.FindCustomer(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetCustomerReturnsNotFound(t *testing.T) {
	token, err := auth.NewToken("customer-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/customers/customer-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.GetCustomer(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetCustomerForbidsAnotherProfile(t *testing.T) {
	token, err := auth.NewToken("customer-2", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{customer: &store.Customer{ID: "customer-1"}}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/customers/customer-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.GetCustomer(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestGetOperatorAllowsOwnProfile(t *testing.T) {
	token, err := auth.NewToken("operator-1", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{operator: &store.Operator{
		ID:      "operator-1",
		Name:    "Operator",
		Email:   "op@example.com",
		License: "LIC-1",
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/operators/operator-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.GetOperator(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestGetOperatorForbidsAnotherProfile(t *testing.T) {
	token, err := auth.NewToken("operator-2", "operator", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	h := NewHandler(&fakeRegistryStore{operator: &store.Operator{ID: "operator-1"}}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/operators/operator-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.GetOperator(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestFindOperatorReturnsOperator(t *testing.T) {
	h := NewHandler(&fakeRegistryStore{operator: &store.Operator{
		ID:    "operator-1",
		Email: "op@example.com",
	}}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/operators?email=op@example.com", nil)
	rec := httptest.NewRecorder()

	h.FindOperator(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestFindOperatorReturnsNotFound(t *testing.T) {
	h := NewHandler(&fakeRegistryStore{}, "test-secret")
	req := httptest.NewRequest(http.MethodGet, "/operators?email=missing@example.com", nil)
	rec := httptest.NewRecorder()

	h.FindOperator(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
