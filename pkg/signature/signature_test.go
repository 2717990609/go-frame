package signature

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	params := map[string]string{"a": "1", "b": "2"}
	secret := "test-secret"
	sig := Generate(params, secret)
	if sig == "" {
		t.Fatal("signature should not be empty")
	}
	if len(sig) != 64 {
		t.Errorf("sha256 hex length should be 64, got %d", len(sig))
	}
}

func TestVerify(t *testing.T) {
	params := map[string]string{"a": "1", "b": "2"}
	secret := "test-secret"
	sig := Generate(params, secret)
	if !Verify(params, sig, secret) {
		t.Fatal("verify should succeed")
	}
	if Verify(params, sig+"x", secret) {
		t.Fatal("verify should fail with wrong sig")
	}
	if Verify(params, sig, "wrong-secret") {
		t.Fatal("verify should fail with wrong secret")
	}
}

func TestGenerateWithNonce(t *testing.T) {
	params := map[string]string{"amount": "100"}
	secret := "key"
	sign, ts, nonce := GenerateWithNonce(params, secret)
	if sign == "" || ts == "" || nonce == "" {
		t.Fatal("all return values should be non-empty")
	}
	params["timestamp"] = ts
	params["nonce"] = nonce
	if !Verify(params, sign, secret) {
		t.Fatal("verify should succeed for GenerateWithNonce output")
	}
}
