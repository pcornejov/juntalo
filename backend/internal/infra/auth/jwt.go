package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTSigner implements app.TokenSigner with short-lived HS256 access tokens
// (Etapa 2 §2.5: 15 min, refresh token separado en cookie httpOnly).
type JWTSigner struct {
	secret []byte
}

func NewJWTSigner(secret string) *JWTSigner {
	return &JWTSigner{secret: []byte(secret)}
}

func (s *JWTSigner) Sign(userID uuid.UUID, ttl time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTSigner) Parse(tokenString string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("jwt: invalid token: %w", err)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("jwt: invalid subject: %w", err)
	}
	return userID, nil
}
