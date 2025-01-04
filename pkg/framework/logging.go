package framework

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type loggerContextKey struct{}

func GetLogger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return newLogger()
	}
	logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger)
	if !ok {
		return newLogger()
	}
	return logger
}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func newLogger() *slog.Logger {
	return otelslog.NewLogger("context")
}

func GetHostID() string {
	// https://man7.org/linux/man-pages/man5/machine-id.5.html
	id, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "unknown"
	}
	return string(id)
}

func NewHTTPTransport(transport http.RoundTripper) http.RoundTripper {
	return &meteredTransport{transport: transport}
}

func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: NewHTTPTransport(http.DefaultTransport),
		Timeout:   timeout,
	}
}

type meteredTransport struct {
	transport http.RoundTripper
}

var (
	_      http.RoundTripper = &meteredTransport{}
	tracer                   = otel.Tracer("")
)

func (t *meteredTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx, span := tracer.Start(
		r.Context(),
		"outboundHttpRequest",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("method", r.Method),
			attribute.String("url", r.URL.String()),
		),
	)
	defer span.End()

	startTime := time.Now()
	res, err := t.transport.RoundTrip(r.WithContext(ctx))
	timeTaken := time.Since(startTime)

	errorContent := ""
	if err != nil {
		errorContent = err.Error()
	} else if res.StatusCode >= 400 {
		resBody := copyResBody(res)
		errorContent = string([]rune(resBody)[0:100])
	}

	span.SetAttributes(attribute.Int("status", res.StatusCode))

	GetLogger(ctx).InfoContext(
		ctx, "outbound HTTP request",
		"method", r.Method,
		"url", r.URL.String(),
		"status", res.StatusCode,
		"timeTaken", timeTaken.Milliseconds(),
		"errorContent", errorContent,
	)

	return res, err
}

// copyResBody returns a copy of the response body as a string.
// The res body is still readable after calling.
func copyResBody(res *http.Response) string {
	var b bytes.Buffer
	_, err := io.Copy(&b, res.Body)
	if err != nil {
		return "unknown"
	}

	err = res.Body.Close()
	if err != nil {
		return "unknown"
	}

	bodyContent := b.Bytes()
	res.Body = io.NopCloser(bytes.NewBuffer(bodyContent))

	return string(bodyContent)
}
