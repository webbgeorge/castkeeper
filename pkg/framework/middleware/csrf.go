package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

type CSRFMiddleware struct {
	CSRFSecretKey    string
	CSRFSecureCookie bool
}

// wraps the gorilla CSRF Middleware
func (mw CSRFMiddleware) Handler(next framework.Handler, _ framework.MiddlewareConfig) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		var err error
		hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = next(r.Context(), w, r)
		})
		csrf.Protect(
			[]byte(mw.CSRFSecretKey),
			csrf.Secure(mw.CSRFSecureCookie),
			csrf.Path("/"),
		)(hf).ServeHTTP(w, r)
		return err
	}
}

func (mw CSRFMiddleware) Match(config framework.MiddlewareConfig) bool {
	return false
}
