// Package googleauth implements app.GoogleTokenVerifier verificando el ID
// token (JWT) que entrega el botón de Google Identity Services directamente
// contra las llaves públicas de Google (JWKS) — sin llamar al endpoint
// /tokeninfo de Google en cada login (ese endpoint es solo para debugging,
// según la propia documentación de Google, y tiene rate limit). Las llaves
// se cachean en memoria y se refrescan cuando aparece un kid desconocido.
package googleauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/pcornejov/juntalo/backend/internal/app"
)

const defaultJWKSURL = "https://www.googleapis.com/oauth2/v3/certs"

// googleIssuers: Google firma sus ID tokens con cualquiera de estos dos
// valores de "iss" indistintamente (ambos documentados) — hay que aceptar
// los dos, no solo uno.
var googleIssuers = map[string]bool{
	"accounts.google.com":         true,
	"https://accounts.google.com": true,
}

type jwk struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type Verifier struct {
	clientID string
	jwksURL  string
	client   *http.Client

	mu   sync.Mutex
	keys map[string]*rsa.PublicKey
}

// NewVerifier arma un verificador para un GOOGLE_CLIENT_ID dado — ese es el
// "audience" que debe traer todo ID token para que Juntalo lo acepte (evita
// que un token válido emitido para OTRA app de Google se pueda reusar acá).
func NewVerifier(clientID string) *Verifier {
	return &Verifier{clientID: clientID, jwksURL: defaultJWKSURL, client: &http.Client{Timeout: 5 * time.Second}}
}

// Verify valida firma, issuer, audience y expiración del ID token, y
// devuelve los claims que el caso de uso necesita para loguear/crear al
// usuario. No valida email_verified acá — es una decisión de negocio (Login
// con Google decide si lo exige), no de criptografía.
func (v *Verifier) Verify(ctx context.Context, idToken string) (app.GoogleClaims, error) {
	token, err := jwt.Parse(idToken, v.keyFunc(ctx),
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithAudience(v.clientID),
	)
	if err != nil {
		return app.GoogleClaims{}, fmt.Errorf("googleauth: verify token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return app.GoogleClaims{}, fmt.Errorf("googleauth: invalid token claims")
	}

	iss, _ := claims["iss"].(string)
	if !googleIssuers[iss] {
		return app.GoogleClaims{}, fmt.Errorf("googleauth: unexpected issuer %q", iss)
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	if sub == "" || email == "" {
		return app.GoogleClaims{}, fmt.Errorf("googleauth: token missing sub/email")
	}
	emailVerified, _ := claims["email_verified"].(bool)
	fullName, _ := claims["name"].(string)

	return app.GoogleClaims{Subject: sub, Email: email, EmailVerified: emailVerified, FullName: fullName}, nil
}

func (v *Verifier) keyFunc(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("googleauth: token missing kid")
		}

		if key := v.cachedKey(kid); key != nil {
			return key, nil
		}
		if err := v.refreshKeys(ctx); err != nil {
			return nil, err
		}
		if key := v.cachedKey(kid); key != nil {
			return key, nil
		}
		return nil, fmt.Errorf("googleauth: unknown key id %q", kid)
	}
}

func (v *Verifier) cachedKey(kid string) *rsa.PublicKey {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.keys[kid]
}

// refreshKeys se llama solo cuando aparece un kid que el cache no tiene
// (rotación de llaves de Google es infrecuente) — no hay TTL propio, el kid
// desconocido ES la señal de que hay que refrescar.
func (v *Verifier) refreshKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("googleauth: build jwks request: %w", err)
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("googleauth: fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	var out jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("googleauth: decode jwks: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(out.Keys))
	for _, k := range out.Keys {
		pub, err := parseRSAPublicKey(k.N, k.E)
		if err != nil {
			continue // llave que no pudimos parsear: se ignora, no aborta el refresh completo
		}
		keys[k.Kid] = pub
	}

	v.mu.Lock()
	v.keys = keys
	v.mu.Unlock()
	return nil
}

func parseRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}
