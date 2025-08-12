package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/framework"
)

// loggingResponseWriter wraps http.ResponseWriter to capture the status code sent to the client
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

type LogMiddleware struct{}

func (mw LogMiddleware) Handler(next framework.Handler, _ framework.MiddlewareConfig) framework.Handler {
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

func (mw LogMiddleware) Match(config framework.MiddlewareConfig) bool {
	return false
}
