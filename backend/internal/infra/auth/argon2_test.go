package auth

import "testing"

func TestArgon2idHasher_HashAndVerify(t *testing.T) {
	h := NewArgon2idHasher()

	hash, err := h.Hash("password123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !h.Verify("password123", hash) {
		t.Error("expected password to verify against its own hash")
	}
	if h.Verify("wrong-password", hash) {
		t.Error("expected wrong password to fail verification")
	}
}

func TestArgon2idHasher_DifferentSaltsPerHash(t *testing.T) {
	h := NewArgon2idHasher()

	h1, _ := h.Hash("password123")
	h2, _ := h.Hash("password123")
	if h1 == h2 {
		t.Error("expected different salts to produce different hashes for the same password")
	}
}
