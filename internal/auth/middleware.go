package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is a private type for this package's context keys.
type contextKey string

const userIDKey contextKey = "userID"

type AuthMiddleware struct {
	tokens *TokenService
}

func NewAuthMiddleware(tokens *TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens}
}

// RequireAuth wraps a handler so it runs only for requests carrying a valid
// access token. On any failure it writes a generic 401
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Expect the header "Authorization: Bearer <token>".
		header := r.Header.Get("Authorization")
		rawToken, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || rawToken == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := m.tokens.VerifyAccessToken(rawToken)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Hand the verified identity to the downstream handler via the context.
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext returns the authenticated user ID stored by RequireAuth.
// The bool is false if the request did not pass through RequireAuth.
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
