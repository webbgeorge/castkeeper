package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

type ErrorHandlerMiddleware struct{}

func (mw ErrorHandlerMiddleware) Handler(next framework.Handler, _ framework.MiddlewareConfig) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		defer func() {
			if rec := recover(); rec != nil {
				framework.GetLogger(ctx).ErrorContext(
					ctx, "recovered from panic",
					"error", fmt.Sprintf("Panic: %+v, req: %s %s", rec, r.Method, r.URL.Path),
				)
				_ = framework.Render(ctx, w, 500, pages.Error("Internal server error"))
			}
		}()

		err := next(ctx, w, r)
		if err != nil {
			var httpError framework.HttpError
			switch err := err.(type) {
			case framework.HttpError:
				// info log expected http errors (i.e. not a 500)
				framework.GetLogger(ctx).InfoContext(
					ctx, "handled HTTP error",
					"error", err.Error(),
				)
				httpError = err
			default:
				framework.GetLogger(ctx).ErrorContext(
					ctx, "unhandled error",
					"error", err.Error(),
				)
				httpError = framework.HttpError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Internal server error",
				}
			}

			if httpError.StatusCode == 404 {
				return framework.Render(ctx, w, httpError.StatusCode, pages.NotFound())
			}
			return framework.Render(ctx, w, httpError.StatusCode, pages.Error(httpError.Message))
		}

		return nil
	}
}

func (mw ErrorHandlerMiddleware) Match(config framework.MiddlewareConfig) bool {
	return false
}
