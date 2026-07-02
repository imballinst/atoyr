package middleware

import (
	"context"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	verifier      *oidc.IDTokenVerifier
	allowedEmails map[string]bool
}

func NewNoopAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func NewAuthMiddleware(allowedEmails []string) (*AuthMiddleware, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
	})

	emailSet := make(map[string]bool)
	for _, email := range allowedEmails {
		emailSet[email] = true
	}

	return &AuthMiddleware{
		verifier:      verifier,
		allowedEmails: emailSet,
	}, nil
}

func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if a.verifier == nil {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization header"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		idToken, err := a.verifier.Verify(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		var claims struct {
			Email string `json:"email"`
		}
		if err := idToken.Claims(&claims); err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid claims"})
			return
		}

		if !a.allowedEmails[claims.Email] {
			c.AbortWithStatusJSON(403, gin.H{"error": "email not allowed"})
			return
		}

		c.Set("userEmail", claims.Email)
		c.Next()
	}
}
