package middleware

import (
	"context"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/framework"
)

type ResponseHeaderMiddleware struct{}

func (mw ResponseHeaderMiddleware) Handler(next framework.Handler, _ framework.MiddlewareConfig) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Embedder-Policy", "credentialless")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' *.mzstatic.com; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'; form-action 'self';")
		return next(ctx, w, r)
	}
}

func (mw ResponseHeaderMiddleware) Match(config framework.MiddlewareConfig) bool {
	return false
}
