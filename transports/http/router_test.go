package http

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPServer(t *testing.T) {
	tt := []struct {
		name         string
		version      string
		idleTimeout  int
		readTimeout  int
		writeTimeout int
		port         int
		route        Route
	}{
		{name: "with api version", version: "v1", port: 8080, route: Route{Name: "test"}},
		{name: "with timeouts", version: "v1", port: 8080, route: Route{Name: "test"}, idleTimeout: 5, writeTimeout: 5, readTimeout: 5},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			s := NewHTTPServer(
				SetServerPort(v.port),
				SetServerApiVersion(v.version),
				SetIdleTimeout(v.idleTimeout),
				SetReadTimeout(v.readTimeout),
				SetWriteTimeout(v.writeTimeout),
			)

			s.Router.Name(v.route.Name)
			path, err := s.Router.Get(v.route.Name).GetPathTemplate()
			if err != nil {
				t.Fatalf("error getting path: %v", err)
			}

			if path != fmt.Sprintf("/api/%s", v.version) {
				t.Errorf("expected %s, but got %s", fmt.Sprintf("/api/%s", v.version), path)
			}

			if s.apiServer.IdleTimeout != time.Duration(v.idleTimeout)*time.Second {
				t.Errorf("expected idle timeout of %v but got %v", time.Duration(v.idleTimeout)*time.Second, s.apiServer.IdleTimeout)
			}

			if s.apiServer.ReadTimeout != time.Duration(v.readTimeout)*time.Second {
				t.Errorf("expected read timeout of %v but got %v", time.Duration(v.readTimeout)*time.Second, s.apiServer.ReadTimeout)
			}

			if s.apiServer.WriteTimeout != time.Duration(v.writeTimeout)*time.Second {
				t.Errorf("expected write timeout of %v but got %v", time.Duration(v.writeTimeout)*time.Second, s.apiServer.WriteTimeout)
			}
		})
	}
}

func TestRegisterSubrouter(t *testing.T) {
	prefix := "/test"
	routes := []Route{
		{
			Name:   "test",
			Method: http.MethodGet,
			Path:   "/test",
		},
	}

	s := NewHTTPServer()

	s.RegisterSubRouter(prefix, routes)

	for _, v := range routes {
		t.Run(v.Name, func(t *testing.T) {
			path, err := s.Router.Get(v.Name).GetPathTemplate()
			if err != nil {
				t.Fatal(err)
			}

			methods, err := s.Router.Get(v.Name).GetMethods()
			if err != nil {
				t.Fatal(err)
			}

			if path != fmt.Sprintf("%s%s", prefix, v.Path) {
				t.Errorf("expected path %s but got %s", fmt.Sprintf("%s%s", prefix, v.Path), path)
			}

			if methods[0] != v.Method {
				t.Errorf("expected method %s but got %s", v.Method, methods[0])
			}
		})
	}
}
