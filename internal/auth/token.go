package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

const tokenTTL = 12 * time.Hour

type Claims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

func NewToken(userID, role, secret string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Role:      role,
		ExpiresAt: time.Now().Add(tokenTTL).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	payloadPart := base64.RawURLEncoding.EncodeToString(payload)
	signature := sign(payloadPart, secret)
	return payloadPart + "." + signature, nil
}

func VerifyToken(token, secret string) (*Claims, bool) {
	payloadPart, signature, ok := strings.Cut(token, ".")
	if !ok || payloadPart == "" || signature == "" {
		return nil, false
	}

	if !hmac.Equal([]byte(signature), []byte(sign(payloadPart, secret))) {
		return nil, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return nil, false
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, false
	}
	if claims.UserID == "" || claims.Role == "" || time.Now().Unix() > claims.ExpiresAt {
		return nil, false
	}

	return &claims, true
}

func sign(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload)) //nolint:errcheck
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
