package logx

import (
	"context"
	"net/http"
)

func SessionLogger(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithCtx(r.Context(), FromCtx(ctx))
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
