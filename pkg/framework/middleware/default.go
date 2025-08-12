package middleware

import (
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

func DefaultMiddlewareStack(
	csrfSecretKey string,
	csrfSecureCookie bool,
) []framework.Middleware {
	return []framework.Middleware{
		LogMiddleware{},
		ResponseHeaderMiddleware{},
		ErrorHandlerMiddleware{},
		CSRFMiddleware{
			CSRFSecretKey:    csrfSecretKey,
			CSRFSecureCookie: csrfSecureCookie,
		},
	}
}
