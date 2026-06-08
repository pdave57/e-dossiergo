// Package token handles JWT creation and verification for e-Dossier.
package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims embedded in every e-Dossier JWT.
type Claims struct {
	UserID   string   `json:"user_id"`
	SchoolID string   `json:"school_id,omitempty"` // empty for state-level staff
	StateID  string   `json:"state_id"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// Maker signs and verifies JWTs.
type Maker struct {
	secret        []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// New creates a Maker.
func New(secret string, accessTTL, refreshTTL time.Duration) *Maker {
	return &Maker{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// CreateAccessToken mints a short-lived access JWT.
func (m *Maker) CreateAccessToken(userID, stateID, schoolID string, roles []string) (string, time.Time, error) {
	exp := time.Now().Add(m.accessTTL)
	claims := Claims{
		UserID:   userID,
		StateID:  stateID,
		SchoolID: schoolID,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    "e-dossier",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(m.secret)
	return signed, exp, err
}

// CreateRefreshToken mints a long-lived refresh JWT (no roles embedded).
func (m *Maker) CreateRefreshToken(userID string) (string, time.Time, error) {
	exp := time.Now().Add(m.refreshTTL)
	claims := jwt.RegisteredClaims{
		ID:        uuid.NewString(),
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(exp),
		Issuer:    "e-dossier-refresh",
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(m.secret)
	return signed, exp, err
}

// Verify parses and validates an access token, returning its Claims.
func (m *Maker) Verify(tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// VerifyRefresh parses a refresh token returning the subject (userID).
func (m *Maker) VerifyRefresh(tokenStr string) (string, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := tok.Claims.(*jwt.RegisteredClaims)
	if !ok || !tok.Valid {
		return "", errors.New("invalid refresh token")
	}
	return claims.Subject, nil
}
