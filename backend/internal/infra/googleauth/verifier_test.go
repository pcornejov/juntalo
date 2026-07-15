package googleauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// newTestServer sirve un JWKS falso (nuestra propia llave RSA) en vez del
// real de Google — determinista, sin red, mismo patrón que
// infra/payments/webpay/webpay_test.go.
func newTestServer(t *testing.T, key *rsa.PrivateKey, kid string) *httptest.Server {
	t.Helper()
	jwks := jwksResponse{Keys: []jwk{{
		Kid: kid,
		N:   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big64(key.PublicKey.E)),
	}}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func big64(e int) []byte {
	// e siempre es 65537 en las llaves de Google — 3 bytes bastan.
	return []byte{byte(e >> 16), byte(e >> 8), byte(e)}
}

func signToken(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func baseClaims(clientID string) jwt.MapClaims {
	return jwt.MapClaims{
		"iss":            "https://accounts.google.com",
		"aud":            clientID,
		"sub":            "1234567890",
		"email":          "vecina@example.com",
		"email_verified": true,
		"name":           "Vecina Prueba",
		"exp":            time.Now().Add(time.Hour).Unix(),
		"iat":            time.Now().Unix(),
	}
}

func TestVerifier_Verify_Success(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestServer(t, key, "kid-1")

	v := NewVerifier("my-client-id")
	v.jwksURL = srv.URL

	tok := signToken(t, key, "kid-1", baseClaims("my-client-id"))

	claims, err := v.Verify(context.Background(), tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Subject != "1234567890" || claims.Email != "vecina@example.com" || !claims.EmailVerified {
		t.Errorf("unexpected claims: %+v", claims)
	}
	if claims.FullName != "Vecina Prueba" {
		t.Errorf("got full name %q", claims.FullName)
	}
}

func TestVerifier_Verify_WrongAudience(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestServer(t, key, "kid-1")

	v := NewVerifier("my-client-id")
	v.jwksURL = srv.URL

	tok := signToken(t, key, "kid-1", baseClaims("someone-elses-client-id"))

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("expected an error for mismatched audience")
	}
}

func TestVerifier_Verify_WrongIssuer(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestServer(t, key, "kid-1")

	v := NewVerifier("my-client-id")
	v.jwksURL = srv.URL

	claims := baseClaims("my-client-id")
	claims["iss"] = "https://evil.example.com"
	tok := signToken(t, key, "kid-1", claims)

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("expected an error for unexpected issuer")
	}
}

func TestVerifier_Verify_UnknownKid(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestServer(t, key, "kid-1")

	v := NewVerifier("my-client-id")
	v.jwksURL = srv.URL

	tok := signToken(t, key, "kid-does-not-exist", baseClaims("my-client-id"))

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("expected an error for an unknown key id")
	}
}

func TestVerifier_Verify_ExpiredToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	srv := newTestServer(t, key, "kid-1")

	v := NewVerifier("my-client-id")
	v.jwksURL = srv.URL

	claims := baseClaims("my-client-id")
	claims["exp"] = time.Now().Add(-time.Hour).Unix()
	tok := signToken(t, key, "kid-1", claims)

	if _, err := v.Verify(context.Background(), tok); err == nil {
		t.Fatal("expected an error for an expired token")
	}
}
