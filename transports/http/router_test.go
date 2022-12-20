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
		{name: "with api version", port: 8080},
		{name: "with timeouts", port: 8080, idleTimeout: 5, writeTimeout: 5, readTimeout: 5},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			s := NewHTTPServer(
				SetServerPort(v.port),
				SetIdleTimeout(v.idleTimeout),
				SetReadTimeout(v.readTimeout),
				SetWriteTimeout(v.writeTimeout),
			)

			if s.apiServer.Addr != fmt.Sprintf(":%d", v.port) {
				t.Errorf("expected port to be %d but got %v", v.port, s.apiServer.Addr)
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
			Method: http.MethodGet,
			Path:   "/test",
		},
	}

	s := NewHTTPServer()

	s.RegisterSubRouter(prefix, routes)

	paths := s.Router.Routes()

	for _, path := range paths {
		if path.Pattern != fmt.Sprintf("%s/*", prefix) {
			t.Errorf("expected prefix %s but got %s", fmt.Sprintf("%s/*", prefix), path)
		}
	}
}
