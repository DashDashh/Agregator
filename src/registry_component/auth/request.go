package auth

import (
	"net/http"
	"strings"
)

type User struct {
	ID   string
	Role string
}

func UserFromRequest(r *http.Request, secret string) (*User, bool) {
	header := r.Header.Get("Authorization")
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	if token == "" || token == header {
		return nil, false
	}

	claims, ok := VerifyToken(token, secret)
	if !ok {
		return nil, false
	}

	return &User{ID: claims.UserID, Role: claims.Role}, true
}
