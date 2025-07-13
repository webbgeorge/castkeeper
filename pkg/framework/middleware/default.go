package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

func DefaultMiddlewareStack() []framework.Middleware {
	return []framework.Middleware{
		NewLogMiddleware(),
		NewResHeaderMiddleware(),
		NewErrorHandlerMiddleware(),
	}
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code sent to the client
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func NewLogMiddleware() framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			startTime := time.Now()

			// inherit server logger
			logger := framework.GetLogger(ctx)
			if logger == nil {
				log.Printf("WARN: no logger in request context (LogMiddleware)")
				return next(ctx, w, r)
			}

			// add the updated logger to the ctx, so it can be used throughout the request
			ctxWithLogger := framework.ContextWithLogger(ctx, logger)

			// captures the status code from the res writer, so we can log it
			loggingResW := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

			err := next(ctxWithLogger, loggingResW, r)

			timeTaken := time.Since(startTime)
			logger.InfoContext(
				ctxWithLogger,
				"HTTP request",
				"path", r.URL.Path,
				"method", r.Method,
				"status", loggingResW.statusCode,
				"timeTaken", timeTaken.Milliseconds(),
			)

			return err
		}
	}
}

func NewResHeaderMiddleware() framework.Middleware {
	return func(next framework.Handler) framework.Handler {
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
}

func NewErrorHandlerMiddleware() framework.Middleware {
	return func(next framework.Handler) framework.Handler {
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
}

// wraps the gorilla CSRF Middleware
func NewCSRFMiddleware(csrfSecretKey string, csrfSecureCookie bool) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var err error
			hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = next(r.Context(), w, r)
			})
			csrf.Protect(
				[]byte(csrfSecretKey),
				csrf.Secure(csrfSecureCookie),
				csrf.Path("/"),
			)(hf).ServeHTTP(w, r)
			return err
		}
	}
}
