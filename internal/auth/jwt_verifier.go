package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingToken = errors.New("missing token")
	ErrInvalidToken = errors.New("invalid token")
)

type Verifier struct {
	jwks      *keyfunc.JWKS
	issuer    string
	clockSkew time.Duration
}

func NewVerifier(jwksURL, issuer string) (*Verifier, error) {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute,
		RefreshTimeout:    10 * time.Second,
		RefreshUnknownKID: true,
	})
	if err != nil {
		return nil, err
	}

	return &Verifier{
		jwks:      jwks,
		issuer:    issuer,
		clockSkew: 30 * time.Second,
	}, nil
}

type CustomClaims struct {
	jwt.RegisteredClaims

	Scope string `json:"scope,omitempty"`
	Azp   string `json:"azp,omitempty"`

	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	GivenName     string `json:"user_first_name,omitempty"`
	FamilyName    string `json:"user_last_name,omitempty"`
	Username      string `json:"username,omitempty"`
	Sid           string `json:"sid,omitempty"`
}

func (v *Verifier) VerifyAccessToken(raw string) (*CustomClaims, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrMissingToken
	}

	claims := new(CustomClaims)

	token, err := jwt.ParseWithClaims(
		raw,
		claims,
		v.jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithLeeway(v.clockSkew),
		jwt.WithIssuer(v.issuer),
	)

	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
