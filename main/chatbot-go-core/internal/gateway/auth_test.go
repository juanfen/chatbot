package gateway

import "testing"

func TestAuthManagerGenerateAndVerify(t *testing.T) {
	auth := NewAuthManager("secret", 3600)
	token := auth.GenerateToken("client-1", "device-1")
	if !auth.VerifyToken(token, "client-1", "device-1") {
		t.Fatal("expected token to verify")
	}
	if auth.VerifyToken(token, "client-2", "device-1") {
		t.Fatal("expected token with wrong client to fail")
	}
}

func TestBearerToken(t *testing.T) {
	if got := BearerToken("Bearer abc"); got != "abc" {
		t.Fatalf("unexpected bearer token: %q", got)
	}
	if got := BearerToken("raw"); got != "raw" {
		t.Fatalf("unexpected raw token: %q", got)
	}
}
