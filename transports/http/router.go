package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CoverWhale/coverwhale-go/logging"
	"github.com/gorilla/mux"
)

var (
	InternalError = fmt.Errorf("internal server error")
)

// ServerOption is a functional option to modify the server
type ServerOption func(*Server)

// handlerWithError is a normal HTTP handler but returns an error
type handlerWithError func(http.ResponseWriter, *http.Request) error

// errHandler contains a handler that returns an error and a logger
type errHandler struct {
	handler handlerWithError
	logger  *logging.Logger
}

// Server holds the http.Server, a logger, and the router to attach to the http.Server
type Server struct {
	apiServer *http.Server
	Logger    *logging.Logger
	Router    *mux.Router
}

// Route contains the information needed for an HTTP handler
type Route struct {
	Name    string
	Method  string
	Path    string
	Handler handlerWithError
}

// NewHTTPServer initializes and returns a new Server
func NewHTTPServer(opts ...ServerOption) *Server {
	r := mux.NewRouter().StrictSlash(true)

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

// SetServerPort sets the server listening port
func SetServerPort(p int) ServerOption {
	return func(s *Server) {
		s.apiServer.Addr = fmt.Sprintf(":%d", p)
	}
}

// SetServerApiVersion sets the default API version path. If this option is not provided then the handlers are registered to /
func SetServerApiVersion(v string) ServerOption {
	return func(s *Server) {
		r := s.Router.PathPrefix(fmt.Sprintf("/api/%s", v)).Subrouter()
		s.Router = r
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
func (e *errHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := e.handler(w, r)
	if err == nil {
		return
	}

	ce, ok := err.(ClientError)
	if !ok {
		e.logger.Errorf("status=%d, err=%v", http.StatusInternalServerError, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(InternalError.Error()))
		return
	}

	body := ce.Body()

	if ce.Status() == 401 || ce.Status() == 403 {
		e.logger.Errorf("staus=%d, err=%s", ce.Status(), ce.Error())
	}

	w.WriteHeader(ce.Status())
	w.Write(body)
}

// RegisterSubRouter creates a subrouter based on a path and a slice of routes
func (s *Server) RegisterSubRouter(prefix string, routes []Route) *Server {
	subRouter := s.Router.PathPrefix(prefix).Subrouter()

	for _, v := range routes {
		subRouter.Methods(v.Method).Path(v.Path).Name(v.Name).Handler(&errHandler{
			handler: v.Handler,
			logger:  s.Logger,
		})
	}

	return s
}

// Serve starts the http.Server
func (s *Server) Serve() error {
	return s.apiServer.ListenAndServe()
}
