package middleware

import "context"

type ctxKey string

const (
	ctxUserID ctxKey = "user_id"
	ctxRole   ctxKey = "role"
)

func withUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, ctxUserID, id)
}
func withRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ctxRole, role)
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(ctxUserID)
	id, ok := v.(int64)
	return id, ok
}

func RoleFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxRole)
	r, ok := v.(string)
	return r, ok
}