// Copyright 2023 Cover Whale Insurance Solutions Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/CoverWhale/logr"
)

var (
	ErrTestingError = fmt.Errorf("testing error")
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
	prefix := "/api/v1/"
	routes := []Route{
		{
			Method: http.MethodGet,
			Path:   "/test",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		},
	}

	s := NewHTTPServer()

	s.RegisterSubRouter(prefix, routes)

	for _, route := range routes {
		if route.Path == "/healthz" {
			continue
		}

		req := &http.Request{
			Method: route.Method,
			URL: &url.URL{
				Path: fmt.Sprintf("%s%s", prefix, route.Path),
			},
		}

		_, pattern := s.Router.Handler(req)
		if pattern != prefix {
			t.Errorf("expected prefix %s but got %s", prefix, pattern)
		}
	}
}

func TestErrHandlerServeHTTP(t *testing.T) {
	tt := []struct {
		name    string
		handler ErrHandler
		err     error
		status  int
	}{
		{
			name: "400 error", handler: ErrHandler{
				Handler: func(w http.ResponseWriter, r *http.Request) error {
					return NewClientError(ErrTestingError, 400)
				},
				Logger: logr.NewLogger(),
			},
			err:    NewClientError(ErrTestingError, 400),
			status: 400,
		},
		{
			name: "500 error", handler: ErrHandler{
				Handler: func(w http.ResponseWriter, r *http.Request) error {
					return ErrInternalError
				},
				Logger: logr.NewLogger(),
			},
			err:    ErrInternalError,
			status: 500,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/testing", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			v.handler.ServeHTTP(rr, req)

			if status := rr.Code; status != v.status {
				t.Errorf("Expected status %d but got %d", v.status, status)
			}
		})
	}

}
