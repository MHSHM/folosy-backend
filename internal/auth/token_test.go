package auth

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestGenerateAccessToken(t *testing.T) {
	svc := NewTokenService("test-secret", 15*time.Minute, 7*24*time.Hour)

	token, err := svc.GenerateAccessToken("user-123")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	// A JWT is three dot-separated parts: header.payload.signature.
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}

	// The payload (part 2) is base64url-encoded JSON. Decode it to peek inside.
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	t.Logf("full token: %s", token)
	t.Logf("decoded payload: %s", payload)

	// The user ID we passed in should appear as the subject claim.
	if !strings.Contains(string(payload), "user-123") {
		t.Errorf("payload missing the user id: %s", payload)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	svc := NewTokenService("test-secret", 15*time.Minute, 7*24*time.Hour)

	rt, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	t.Logf("raw (to client):  %s", rt.Raw)
	t.Logf("hash (to db):     %s", rt.Hash)
	t.Logf("expires at:       %s", rt.ExpiresAt)

	if rt.Raw == rt.Hash {
		t.Error("hash equals raw token; it must be hashed before storage")
	}

	if len(rt.Hash) != 64 {
		t.Errorf("expected 64-char sha256 hex, got %d chars", len(rt.Hash))
	}

	if !rt.ExpiresAt.After(time.Now()) {
		t.Error("expiry should be in the future")
	}
}
