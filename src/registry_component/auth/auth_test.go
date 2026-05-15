package auth

import (
	"net/http/httptest"
	"testing"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("strongpass123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" || hash == "strongpass123" {
		t.Fatalf("HashPassword returned unsafe hash %q", hash)
	}
	if !VerifyPassword("strongpass123", hash) {
		t.Fatal("VerifyPassword rejected the original password")
	}
	if VerifyPassword("wrong-password", hash) {
		t.Fatal("VerifyPassword accepted a wrong password")
	}
}

func TestTokenRoundTrip(t *testing.T) {
	token, err := NewToken("user-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}

	claims, ok := VerifyToken(token, "test-secret")
	if !ok {
		t.Fatal("VerifyToken rejected a token signed with the same secret")
	}
	if claims.UserID != "user-1" || claims.Role != "customer" {
		t.Fatalf("claims = %+v, want user-1/customer", claims)
	}
	if _, ok := VerifyToken(token, "other-secret"); ok {
		t.Fatal("VerifyToken accepted a token signed with another secret")
	}
}

func TestUserFromRequest(t *testing.T) {
	token, err := NewToken("user-1", "customer", "test-secret")
	if err != nil {
		t.Fatalf("NewToken returned error: %v", err)
	}
	req := httptest.NewRequest("GET", "/orders", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	user, ok := UserFromRequest(req, "test-secret")
	if !ok {
		t.Fatal("UserFromRequest rejected a valid bearer token")
	}
	if user.ID != "user-1" || user.Role != "customer" {
		t.Fatalf("user = %+v, want user-1/customer", user)
	}
}

func TestUserFromRequestRejectsMissingOrMalformedToken(t *testing.T) {
	req := httptest.NewRequest("GET", "/orders", nil)
	if _, ok := UserFromRequest(req, "test-secret"); ok {
		t.Fatal("UserFromRequest accepted missing Authorization header")
	}

	req.Header.Set("Authorization", "Basic token")
	if _, ok := UserFromRequest(req, "test-secret"); ok {
		t.Fatal("UserFromRequest accepted non-bearer Authorization header")
	}
}
