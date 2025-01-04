package framework

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/components/pages"
)

type Middleware func(next Handler) Handler

func DefaultMiddlewareStack() []Middleware {
	return []Middleware{
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

func NewLogMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			startTime := time.Now()

			// inherit server logger
			logger := GetLogger(ctx)
			if logger == nil {
				log.Printf("WARN: no logger in request context (LogMiddleware)")
				return next(ctx, w, r)
			}

			// add the updated logger to the ctx, so it can be used throughout the request
			ctxWithLogger := ContextWithLogger(ctx, logger)

			// captures the status code from the res writer, so we can log it
			loggingResW := &loggingResponseWriter{ResponseWriter: w, statusCode: 200}

			err := next(ctxWithLogger, loggingResW, r)

			timeTaken := time.Now().Sub(startTime)
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

func NewResHeaderMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
			w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
			w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; connect-src 'self'; img-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'; form-action 'self';")
			return next(ctx, w, r)
		}
	}
}

func NewErrorHandlerMiddleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			defer func() {
				if rec := recover(); rec != nil {
					GetLogger(ctx).ErrorContext(
						ctx, "recovered from panic",
						"error", fmt.Sprintf("Panic: %+v, req: %s %s", rec, r.Method, r.URL.Path),
					)
					_ = Render(ctx, w, 500, pages.Error("Internal server error"))
				}
			}()

			err := next(ctx, w, r)
			if err != nil {
				var httpError HttpError
				switch err := err.(type) {
				case HttpError:
					// info log expected http errors (i.e. not a 500)
					GetLogger(ctx).InfoContext(
						ctx, "handled HTTP error",
						"error", err.Error(),
					)
					httpError = err
				default:
					GetLogger(ctx).ErrorContext(
						ctx, "unhandled error",
						"error", err.Error(),
					)
					httpError = HttpError{
						StatusCode: http.StatusInternalServerError,
						Message:    "Internal server error",
					}
				}

				return Render(ctx, w, httpError.StatusCode, pages.Error(httpError.Message))
			}

			return nil
		}
	}
}
