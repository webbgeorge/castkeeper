package framework

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/a-h/templ"
)

type Server struct {
	Logger *slog.Logger
	addr   string
	Mux    *http.ServeMux
	mws    []Middleware
}

func NewServer(addr string, logger *slog.Logger) *Server {
	sm := http.NewServeMux()
	return &Server{
		Logger: logger,
		addr:   addr,
		Mux:    sm,
	}
}

func (s *Server) SetServerMiddlewares(middlewares ...Middleware) *Server {
	s.mws = middlewares
	return s
}

func (s *Server) AddRoute(pattern string, handler Handler, configs ...MiddlewareConfig) *Server {
	// mws executed in order of slice
	mws := make([]Middleware, 0) // new slice to avoid modifying the order of the args
	mws = append(mws, s.mws...)
	slices.Reverse(mws)

	h := handler
	for _, mw := range mws {
		var mwCfg MiddlewareConfig
		for _, cfg := range configs {
			if ok := mw.Match(cfg); ok {
				mwCfg = cfg
				break
			}
		}
		h = mw.Handler(h, mwCfg)
	}

	handlerFunc := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := ContextWithLogger(r.Context(), s.Logger)
			err := h(ctx, w, r)
			if err != nil {
				// errors should have been handled by middleware
				s.Logger.ErrorContext(ctx, fmt.Sprintf("Unhandled error: %s\n", err))
			}
		},
	)
	s.Mux.Handle(pattern, handlerFunc)

	return s
}

func (s *Server) AddFileServer(path string, fileServer http.Handler, configs ...MiddlewareConfig) *Server {
	s.AddRoute(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fileServer.ServeHTTP(w, r)
		return nil
	}, configs...)
	return s
}

func (s *Server) Start(ctx context.Context) error {
	httpServer := &http.Server{
		Addr:         s.addr,
		Handler:      s.Mux,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Minute * 2,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
	}

	srvErr := make(chan error, 1)
	go func() {
		ln, err := net.Listen("tcp", httpServer.Addr)
		if err != nil {
			srvErr <- err
			return
		}
		s.Logger.Info(fmt.Sprintf("Server listening at '%s'", httpServer.Addr))
		srvErr <- httpServer.Serve(ln)
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

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type MiddlewareConfig any

type Middleware interface {
	Handler(next Handler, config MiddlewareConfig) Handler
	Match(config MiddlewareConfig) bool
}

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

func HttpUnauthorized() HttpError {
	return HttpError{
		StatusCode: 401,
		Message:    "Unauthorized",
	}
}

func HttpForbidden() HttpError {
	return HttpError{
		StatusCode: 403,
		Message:    "Forbidden",
	}
}
