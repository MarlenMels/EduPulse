package middleware

import (
	"net/http"
	"strings"

	"edupulse/internal/auth"
)

func AuthJWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
				writeError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			tokenStr := strings.TrimSpace(h[len("Bearer "):])
			claims, err := auth.ParseToken(tokenStr, secret)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := withUserID(r.Context(), claims.UserID)
			ctx = withRole(ctx, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}