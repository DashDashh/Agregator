package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kirilltahmazidi/aggregator/src/registry_component/auth"
	"github.com/kirilltahmazidi/aggregator/src/shared/httpx"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role     string `json:"role"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}

	req.Role = strings.TrimSpace(req.Role)
	req.Email = strings.TrimSpace(req.Email)
	if req.Role != "customer" && req.Role != "operator" {
		httpx.RespondError(w, http.StatusBadRequest, "роль должна быть customer или operator")
		return
	}
	if req.Email == "" || req.Password == "" {
		httpx.RespondError(w, http.StatusBadRequest, "email и password обязательны")
		return
	}

	var user interface{}
	var userID, passwordHash string
	if req.Role == "customer" {
		c, ok := h.store.GetCustomerByEmail(req.Email)
		if !ok {
			httpx.RespondError(w, http.StatusUnauthorized, "неверный email или пароль")
			return
		}
		user = publicCustomer(c)
		userID = c.ID
		passwordHash = c.PasswordHash
	} else {
		op, ok := h.store.GetOperatorByEmail(req.Email)
		if !ok {
			httpx.RespondError(w, http.StatusUnauthorized, "неверный email или пароль")
			return
		}
		user = publicOperator(op)
		userID = op.ID
		passwordHash = op.PasswordHash
	}

	if !auth.VerifyPassword(req.Password, passwordHash) {
		httpx.RespondError(w, http.StatusUnauthorized, "неверный email или пароль")
		return
	}

	token, err := auth.NewToken(userID, req.Role, h.authSecret)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "ошибка создания токена")
		return
	}

	httpx.Respond(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
		"role":  req.Role,
	})
}
