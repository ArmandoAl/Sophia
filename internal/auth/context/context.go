package context

import (
	"context"
)

type key string

const claimsKey key = "auth_claims"

type Claims struct {
	UserID string
	Email  string
	Role   string
}

func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(Claims)
	return claims, ok
}
