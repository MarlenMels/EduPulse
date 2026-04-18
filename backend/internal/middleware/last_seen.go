package middleware

import (
	"net/http"

	"edupulse/internal/repo"
)

func LastSeen(users *repo.UserRepo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if uid, ok := UserIDFromContext(r.Context()); ok {
				users.UpdateLastSeen(r.Context(), uid)
			}
			next.ServeHTTP(w, r)
		})
	}
}
