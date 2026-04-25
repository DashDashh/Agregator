package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kirilltahmazidi/aggregator/internal/auth"
)

type authUser struct {
	ID   string
	Role string
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role     string `json:"role"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}

	req.Role = strings.TrimSpace(req.Role)
	req.Email = strings.TrimSpace(req.Email)
	if req.Role != "customer" && req.Role != "operator" {
		respondError(w, http.StatusBadRequest, "роль должна быть customer или operator")
		return
	}
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email и password обязательны")
		return
	}

	var user interface{}
	var userID, passwordHash string
	if req.Role == "customer" {
		c, ok := h.store.GetCustomerByEmail(req.Email)
		if !ok {
			respondError(w, http.StatusUnauthorized, "неверный email или пароль")
			return
		}
		user = publicCustomer(c)
		userID = c.ID
		passwordHash = c.PasswordHash
	} else {
		op, ok := h.store.GetOperatorByEmail(req.Email)
		if !ok {
			respondError(w, http.StatusUnauthorized, "неверный email или пароль")
			return
		}
		user = publicOperator(op)
		userID = op.ID
		passwordHash = op.PasswordHash
	}

	if !auth.VerifyPassword(req.Password, passwordHash) {
		respondError(w, http.StatusUnauthorized, "неверный email или пароль")
		return
	}

	token, err := auth.NewToken(userID, req.Role, h.authSecret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "ошибка создания токена")
		return
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
		"role":  req.Role,
	})
}

func (h *Handler) requireAuth(w http.ResponseWriter, r *http.Request) (*authUser, bool) {
	header := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" || token == header {
		respondError(w, http.StatusUnauthorized, "нужна авторизация")
		return nil, false
	}

	claims, ok := auth.VerifyToken(token, h.authSecret)
	if !ok {
		respondError(w, http.StatusUnauthorized, "сессия недействительна или истекла")
		return nil, false
	}

	return &authUser{ID: claims.UserID, Role: claims.Role}, true
}
