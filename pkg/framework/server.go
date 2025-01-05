package framework

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"

	"github.com/a-h/templ"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Server struct {
	Logger       *slog.Logger
	addr         string
	mux          *http.ServeMux
	mws          []Middleware
	onShutdown   func()
	otelResource *resource.Resource
}

func NewServer(otelResource *resource.Resource, addr string) (*Server, error) {
	sm := http.NewServeMux()
	logger := newLogger()
	return &Server{
		Logger: logger,
		addr:   addr,
		mux:    sm,
	}, nil
}

func (s *Server) SetServerMiddlewares(middlewares ...Middleware) *Server {
	s.mws = middlewares
	return s
}

func (s *Server) AddRoute(pattern string, handler Handler, middlewares ...Middleware) *Server {
	// global mws are outermost, executed in order of slice
	mws := make([]Middleware, 0) // new slice to avoid modifying the order of the args
	mws = append(mws, s.mws...)
	mws = append(mws, middlewares...)
	slices.Reverse(mws)

	h := handler
	for _, mw := range mws {
		h = mw(h)
	}

	otelWrappedHandler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := ContextWithLogger(r.Context(), s.Logger)
			err := h(ctx, w, r)
			if err != nil {
				// errors should have been handled by middleware
				s.Logger.ErrorContext(ctx, fmt.Sprintf("Unhandled error: %s\n", err))
			}
		},
	))
	s.mux.Handle(pattern, otelWrappedHandler)

	return s
}

func (s *Server) AddFileServer(path string, fileServer http.Handler, middlewares ...Middleware) *Server {
	s.AddRoute(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fileServer.ServeHTTP(w, r)
		return nil
	}, middlewares...)
	return s
}

func (s *Server) Start(ctx context.Context) error {
	ctx, err := s.initOtel(ctx, s.otelResource)
	if err != nil {
		return err
	}
	defer s.onShutdown()

	httpServer := &http.Server{
		Addr:        s.addr,
		Handler:     otelhttp.NewHandler(s.mux, "web"),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	srvErr := make(chan error, 1)
	go func() {
		s.Logger.Info(fmt.Sprintf("Starting server at '%s'", httpServer.Addr))
		srvErr <- httpServer.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err := <-srvErr:
		// Error when starting HTTP server.
		return fmt.Errorf("failed to start web server: %w", err)
	case <-ctx.Done():
		// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
		err := httpServer.Shutdown(context.Background())
		if err != nil {
			return fmt.Errorf("error when shutting down web server: %w", err)
		}
		s.Logger.Info("Server stopped")
	}

	return nil
}

func (s *Server) initOtel(ctx context.Context, otelResource *resource.Resource) (context.Context, error) {
	ctx, stopSig := signal.NotifyContext(ctx, os.Interrupt)

	otelShutdown, err := setupOTelSDK(ctx, otelResource)
	if err != nil {
		s.Logger.Error(
			"Failed to setup OTel SDK",
			"error", err.Error(),
		)
		return ctx, err
	}

	s.onShutdown = func() {
		stopSig()
		err := otelShutdown(context.Background())
		if err != nil {
			s.Logger.Error(
				"Failed to shutdown OTel",
				"error", err.Error(),
			)
		}
	}

	return ctx, nil
}

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

func Render(ctx context.Context, w http.ResponseWriter, statusCode int, component templ.Component) error {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(statusCode)
	return component.Render(ctx, w)
}

type HttpError struct {
	StatusCode int
	Message    string
}

func (err HttpError) Error() string {
	return fmt.Sprintf("HTTP Error (%d): %s", err.StatusCode, err.Message)
}

func HttpNotFound() HttpError {
	return HttpError{
		StatusCode: 404,
		Message:    "Not found",
	}
}

func HttpBadRequest(msg string) HttpError {
	return HttpError{
		StatusCode: 400,
		Message:    msg,
	}
}
