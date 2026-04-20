package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"crypto/pbkdf2"
)

const (
	hashIterations = 210000
	saltSize       = 16
	keySize        = 32
	tokenTTL       = 12 * time.Hour
)

type Claims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("пароль должен быть не короче 8 символов")
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("salt: %w", err)
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, hashIterations, keySize)
	if err != nil {
		return "", fmt.Errorf("pbkdf2: %w", err)
	}

	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", hashIterations, hex.EncodeToString(salt), hex.EncodeToString(key)), nil
}

func VerifyPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}

	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}

	salt, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}

	expected, err := hex.DecodeString(parts[3])
	if err != nil {
		return false
	}

	actual, err := pbkdf2.Key(sha256.New, password, salt, iter, len(expected))
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(actual, expected) == 1
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
