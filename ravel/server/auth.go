package server

import (
	"crypto/subtle"

	"github.com/danielgtaylor/huma/v2"
)

func unauthorized(ctx huma.Context) {
	ctx.SetStatus(401)
	ctx.BodyWriter().Write([]byte("Unauthorized"))
}

func newAuthMiddleware(bearer []byte) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if bearer == nil {
			next(ctx)
			return
		}

		authorization := ctx.Header("Authorization")
		if authorization == "" {
			unauthorized(ctx)
			return
		}

		result := subtle.ConstantTimeCompare([]byte(authorization), bearer)
		if result != 1 {
			unauthorized(ctx)
			return
		}
		next(ctx)
	}
}
