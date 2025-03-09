package framework

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type loggerContextKey struct{}

func GetLogger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return NewLogger("unknown", "unknown", "unknown", slog.LevelInfo)
	}
	logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger)
	if !ok {
		return NewLogger("unknown", "unknown", "unknown", slog.LevelInfo)
	}
	return logger
}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func NewLogger(appName, envName, version string, logLevel slog.Level) *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	logger = logger.With(
		"app", appName,
		"env", envName,
		"version", version,
		"host", GetHostID(),
	)
	return logger
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

var _ http.RoundTripper = &meteredTransport{}

func (t *meteredTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	startTime := time.Now()
	res, err := t.transport.RoundTrip(r.WithContext(ctx))
	timeTaken := time.Since(startTime)

	errorContent := ""
	statusCode := 0
	if err != nil {
		errorContent = err.Error()
	} else if res.StatusCode >= 400 {
		resBody := copyResBody(res)
		if len(resBody) <= 100 {
			errorContent = resBody
		} else {
			errorContent = resBody[0:100]
		}
	}

	if res != nil {
		statusCode = res.StatusCode
	}

	GetLogger(ctx).InfoContext(
		ctx, "outbound HTTP request",
		"method", r.Method,
		"url", r.URL.String(),
		"status", statusCode,
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
