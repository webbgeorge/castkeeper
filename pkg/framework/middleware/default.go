package middleware

import (
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

func DefaultMiddlewareStack() []framework.Middleware {
	return []framework.Middleware{
		LogMiddleware{},
		ResponseHeaderMiddleware{},
		ErrorHandlerMiddleware{},
	}
}
