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

// RegisterOperator handles POST /operators.
func (h *Handler) RegisterOperator(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		License  string `json:"license"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверное тело запроса: "+err.Error())
		return
	}
	if req.Name == "" || req.License == "" || req.Email == "" || req.Password == "" {
		httpx.RespondError(w, http.StatusBadRequest, "name, license, email и password обязательны")
		return
	}
	if existing, ok := h.store.GetOperatorByEmail(req.Email); ok {
		if existing.PasswordHash == "" {
			passwordHash, err := auth.HashPassword(req.Password)
			if err != nil {
				httpx.RespondError(w, http.StatusBadRequest, err.Error())
				return
			}
			if !h.store.SetOperatorPasswordHash(existing.ID, passwordHash) {
				httpx.RespondError(w, http.StatusInternalServerError, "ошибка сохранения пароля")
				return
			}
			existing.PasswordHash = passwordHash
		}
		if !auth.VerifyPassword(req.Password, existing.PasswordHash) {
			httpx.RespondError(w, http.StatusConflict, "email уже зарегистрирован")
			return
		}
		token, err := auth.NewToken(existing.ID, "operator", h.authSecret)
		if err != nil {
			httpx.RespondError(w, http.StatusInternalServerError, "ошибка создания токена")
			return
		}
		httpx.Respond(w, http.StatusOK, map[string]interface{}{
			"token":    token,
			"user":     publicOperator(existing),
			"role":     "operator",
			"existing": true,
		})
		return
	}
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	op := &store.Operator{
		ID:           uuid.NewString(),
		Name:         req.Name,
		License:      req.License,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}
	if err := h.store.SaveOperator(op); err != nil {
		log.Printf("[api] failed to save operator: %v", err)
		httpx.RespondError(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	log.Printf("[api] operator registered id=%s name=%s", op.ID, op.Name)

	token, err := auth.NewToken(op.ID, "operator", h.authSecret)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "ошибка создания токена")
		return
	}
	httpx.Respond(w, http.StatusCreated, map[string]interface{}{
		"token": token,
		"user":  publicOperator(op),
		"role":  "operator",
	})
}

// FindOperator handles GET /operators?email=...
func (h *Handler) FindOperator(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	if email == "" {
		httpx.RespondError(w, http.StatusBadRequest, "email эксплуатанта не указан")
		return
	}

	op, ok := h.store.GetOperatorByEmail(email)
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, "эксплуатант не найден")
		return
	}

	httpx.Respond(w, http.StatusOK, op)
}

// GetOperator handles GET /operators/{id}.
func (h *Handler) GetOperator(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireAuth(w, r)
	if !ok {
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/operators/")
	if id == "" {
		httpx.RespondError(w, http.StatusBadRequest, "id эксплуатанта не указан")
		return
	}
	if user.Role != "operator" || user.ID != id {
		httpx.RespondError(w, http.StatusForbidden, "нельзя смотреть чужой профиль")
		return
	}

	op, ok := h.store.GetOperator(id)
	if !ok {
		httpx.RespondError(w, http.StatusNotFound, "эксплуатант не найден")
		return
	}

	httpx.Respond(w, http.StatusOK, op)
}

func publicOperator(op *store.Operator) map[string]string {
	return map[string]string{
		"id":      op.ID,
		"name":    op.Name,
		"license": op.License,
		"email":   op.Email,
	}
}
