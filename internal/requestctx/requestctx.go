package requestctx

import (
	"context"
)

type ctxKeyClaims struct{}

type Claims struct {
	Subject           string
	Issuer            string
	Email             string
	EmailVerified     bool
	Name              string
	PreferredUsername string
	GivenName         string
	FamilyName        string
	ExpUnix           int64
}

/*
type ClerkClaims struct {
	jwt.RegisteredClaims

	Sid string `json:"sid,omitempty"` // session id

	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`
	FirstName     string `json:"user_first_name,omitempty"`
	LastName      string `json:"user_last_name,omitempty"`
	Username      string `json:"username,omitempty"`
}
*/

func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, ctxKeyClaims{}, c)
}

func ClaimsFrom(ctx context.Context) (Claims, bool) {
	v := ctx.Value(ctxKeyClaims{})
	c, ok := v.(Claims)
	return c, ok
}
