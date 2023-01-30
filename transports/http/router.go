package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CoverWhale/coverwhale-go/logging"
	"github.com/go-chi/chi/v5"
)

var (
	ErrInternalError = fmt.Errorf("internal server error")
)

// ServerOption is a functional option to modify the server
type ServerOption func(*Server)

type MiddlewareOption func(*chi.Mux)

// handlerWithError is a normal HTTP handler but returns an error
type handlerWithError func(http.ResponseWriter, *http.Request) error

type MiddlewareWithLogger func(*Server, http.Handler) http.Handler

// errHandler contains a handler that returns an error and a logger
type ErrHandler struct {
	Handler handlerWithError
	Logger  *logging.Logger
}

// Server holds the http.Server, a logger, and the router to attach to the http.Server
type Server struct {
	apiServer *http.Server
	Logger    *logging.Logger
	Router    *chi.Mux
}

// Route contains the information needed for an HTTP handler
type Route struct {
	Method  string
	Path    string
	Handler http.Handler
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// NewHTTPServer initializes and returns a new Server
func NewHTTPServer(opts ...ServerOption) *Server {
	r := chi.NewRouter()
	r.Get("/healthz", healthz)

	s := &Server{
		Logger: logging.NewLogger(),
		Router: r,
		apiServer: &http.Server{
			Addr:         ":8080",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.apiServer.Handler = r

	return s
}

func HandleWithContext[T any](h func(http.ResponseWriter, *http.Request, T), ctx T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r, ctx)
	}
}

// SetServerPort sets the server listening port
func SetServerPort(p int) ServerOption {
	return func(s *Server) {
		s.apiServer.Addr = fmt.Sprintf(":%d", p)
	}
}

// SetReadTimeout sets the http.Server read timeout
func SetReadTimeout(t int) ServerOption {
	return func(s *Server) {
		s.apiServer.ReadTimeout = time.Duration(t) * time.Second
	}
}

// SetWriteTimeout sets the http.Server write timeout
func SetWriteTimeout(t int) ServerOption {
	return func(s *Server) {
		s.apiServer.WriteTimeout = time.Duration(t) * time.Second
	}
}

// SetIdleTimeout sets the http.Server idle timeout
func SetIdleTimeout(t int) ServerOption {
	return func(s *Server) {
		s.apiServer.IdleTimeout = time.Duration(t) * time.Second
	}
}

// ServeHTTP satisfies the http.Handler interface to allow for handling of errors from handlers in one place
func (e *ErrHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := e.Handler(w, r)
	if err == nil {
		return
	}

	var ce *ClientError
	if errors.As(err, &ce) {
		w.WriteHeader(ce.Status)
		w.Write([]byte(ce.Body()))
		return
	}

	e.Logger.Errorf("status=%d, err=%v", http.StatusInternalServerError, err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(ErrInternalError.Error()))
}

// RegisterSubRouter creates a subrouter based on a path and a slice of routes. Any middlewares passed in will be mounted to the sub router
func (s *Server) RegisterSubRouter(prefix string, routes []Route, middleware ...func(http.Handler) http.Handler) *Server {
	subRouter := chi.NewRouter()
	s.Router.Mount(prefix, subRouter)
	for _, m := range middleware {
		subRouter.Use(m)
	}

	for _, v := range routes {
		subRouter.Method(v.Method, v.Path, v.Handler)
	}

	return s
}

// Serve starts the http.Server
func (s *Server) Serve(errChan chan<- error) {
	s.Logger.Infof("starting HTTP server on %s", s.apiServer.Addr)
	if err := s.apiServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- err
	}
}

func (s *Server) ShutdownServer(ctx context.Context) {
	s.Logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.apiServer.Shutdown(ctx); err != nil {
		s.Logger.Errorf("error shutting down server: %v", err)
	}

	s.Logger.Info("server stopped")
}
