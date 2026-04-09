package middleware

import "net/http"

func RBAC(allowed ...string) func(http.Handler) http.Handler {
	allowedSet := map[string]struct{}{}
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := RoleFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "missing auth context")
				return
			}
			if _, ok := allowedSet[role]; !ok {
				writeError(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}