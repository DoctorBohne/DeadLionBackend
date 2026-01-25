package requestctx

import "context"

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
	Scope             string
}

func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, ctxKeyClaims{}, c)
}

func ClaimsFrom(ctx context.Context) (Claims, bool) {
	v := ctx.Value(ctxKeyClaims{})
	c, ok := v.(Claims)
	return c, ok
}
