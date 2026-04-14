package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/carlosEA28/ai-code-review/internal/service"
)

type contextKey string

const tokenClaimsContextKey contextKey = "tokenClaims"

func AuthMiddleware(tokenService service.TokenUseCase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tokenService == nil {
				http.Error(w, "authentication service unavailable", http.StatusInternalServerError)
				return
			}

			token := bearerTokenFromHeader(r.Header.Get("Authorization"))
			if token == "" {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}

			claims, err := tokenService.ParseToken(token)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), tokenClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TokenClaimsFromContext(ctx context.Context) (*service.TokenClaims, bool) {
	claims, ok := ctx.Value(tokenClaimsContextKey).(*service.TokenClaims)
	if !ok || claims == nil {
		return nil, false
	}

	return claims, true
}

func bearerTokenFromHeader(headerValue string) string {
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(headerValue, bearerPrefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(headerValue, bearerPrefix))
}
