package middleware

import (
	"net/http"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/auth"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/gin-gonic/gin"
)

const GinClaimsKey = "auth.claims"

func AuthMiddleware(v *auth.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := bearerToken(c.GetHeader("Authorization"))
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Bearer Token"})
			return
		}
		jwtClaims, err := v.VerifyAccessToken(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		expUnix := int64(0)
		if jwtClaims.ExpiresAt != nil {
			expUnix = jwtClaims.ExpiresAt.Time.Unix()
		}
		rc := requestctx.Claims{
			Subject:           jwtClaims.Subject,
			Issuer:            jwtClaims.Issuer,
			ExpUnix:           expUnix,
			EmailVerified:     jwtClaims.EmailVerified,
			Name:              jwtClaims.Name,
			PreferredUsername: jwtClaims.PreferredUsername,
			Email:             jwtClaims.Email,
			GivenName:         jwtClaims.GivenName,
			FamilyName:        jwtClaims.FamilyName,
		}
		ctx := requestctx.WithClaims(c.Request.Context(), rc)
		c.Request = c.Request.WithContext(ctx)
		c.Set("issuer", rc.Issuer)
		c.Set("sub", rc.Subject)
		c.Set(GinClaimsKey, rc)
		c.Next()
	}
}

func bearerToken(h string) string {
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func ClaimsFromGin(c *gin.Context) (requestctx.Claims, bool) {
	v, ok := c.Get(GinClaimsKey)
	if !ok {
		return requestctx.Claims{}, false
	}
	cl, ok := v.(requestctx.Claims)
	return cl, ok
}
