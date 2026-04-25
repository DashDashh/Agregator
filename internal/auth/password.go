package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"crypto/pbkdf2"
)

const (
	hashIterations = 210000
	saltSize       = 16
	keySize        = 32
)

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
