package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
	"github.com/kirilltahmazidi/aggregator/src/shared/store"
)

// RegisterCustomer handles POST /customers.
func (h *Handler) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		httpx.RespondError(w, http.StatusBadRequest, "name, email и password обязательны")
		return
	}
	if existing, ok := h.store.GetCustomerByEmail(req.Email); ok {
		if existing.PasswordHash == "" {
			passwordHash, err := auth.HashPassword(req.Password)
			if err != nil {
				httpx.RespondError(w, http.StatusBadRequest, err.Error())
				return
			}
			if !h.store.SetCustomerPasswordHash(existing.ID, passwordHash) {
				httpx.RespondError(w, http.StatusInternalServerError, "ошибка сохранения пароля")
				return
			}
			existing.PasswordHash = passwordHash
		}
		if !auth.VerifyPassword(req.Password, existing.PasswordHash) {
			httpx.RespondError(w, http.StatusConflict, "email уже зарегистрирован")
			return
		}
		token, err := auth.NewToken(existing.ID, "customer", h.authSecret)
		if err != nil {
			httpx.RespondError(w, http.StatusInternalServerError, "ошибка создания токена")
			return
		}
		httpx.Respond(w, http.StatusOK, map[string]interface{}{
			"token":    token,
			"user":     publicCustomer(existing),
			"role":     "customer",
			"existing": true,
		})
		return
	}
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	c := &store.Customer{
		ID:           uuid.NewString(),
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: passwordHash,
	}
	if err := h.store.SaveCustomer(c); err != nil {
		log.Printf("[api] failed to save customer: %v", err)
		httpx.RespondError(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	log.Printf("[api] customer registered id=%s name=%s", c.ID, c.Name)

	token, err := auth.NewToken(c.ID, "customer", h.authSecret)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "ошибка создания токена")
		return
	}
	httpx.Respond(w, http.StatusCreated, map[string]interface{}{
		"token": token,
		"user":  publicCustomer(c),
		"role":  "customer",
	})
}

// FindCustomer handles GET /customers?email=...
func (h *Handler) FindCustomer(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	if email == "" {
		httpx.RespondError(w, http.StatusBadRequest, "email заказчика не указан")
		return
	}

	c, ok := h.store.GetCustomerByEmail(email)
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, "заказчик не найден")
		return
	}

	httpx.Respond(w, http.StatusOK, c)
}

// GetCustomer handles GET /customers/{id}.
func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/customers/")
	if path == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id заказчика не указан")
		return
	}
	if user.Role != "customer" || user.ID != path {
		httpx.RespondError(w, http.StatusForbidden, "нельзя смотреть чужой профиль")
		return
	}

	c, ok := h.store.GetCustomer(path)
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, "заказчик не найден")
		return
	}

	httpx.Respond(w, http.StatusOK, c)
}

func publicCustomer(c *store.Customer) map[string]string {
	return map[string]string{
		"id":    c.ID,
		"name":  c.Name,
		"email": c.Email,
		"phone": c.Phone,
	}
}
