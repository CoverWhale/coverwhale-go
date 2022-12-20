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

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/CoverWhale/logr"
)

type DataStore interface {
	GetData(id string) string
}

type Server interface {
	Serve(chan<- error)
	ShutdownServer(context.Context)
}

type App struct {
	Server Server
	DS     DataStore
}

type InMem struct {
	Data map[string]string
}

func (i InMem) GetData(id string) string {
	return i.Data[id]
}

func (a *App) getSampleRoutes() []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method:  http.MethodGet,
			Path:    "/testing",
			Handler: http.HandlerFunc(a.exampleGet),
		},
		{
			Method:  http.MethodGet,
			Path:    "/testingCustom",
			Handler: cwhttp.HandleWithContext(customHandlerType, a.DS),
		},
	}
}

func (a *App) exampleGet(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	data := a.DS.GetData(id)

	w.Write([]byte(data))
}

func customHandlerType(w http.ResponseWriter, r *http.Request, ds DataStore) {
	id := r.URL.Query().Get("id")

	data := ds.GetData(id)

	w.Write([]byte(data))
}

func exampleMiddleware(l *logr.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				l.Info("unauthorized")
				w.WriteHeader(401)
				w.Write([]byte("unauthorized"))
				return
			}

			l.Info("in middleware")
			h.ServeHTTP(w, r)
		})
	}

}

func main() {
	ctx := context.Background()
	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
	)
	ds := InMem{
		Data: map[string]string{
			"test": "testing",
		},
	}

	a := App{
		Server: s,
		DS:     ds,
	}

	s.RegisterSubRouter("/api/v1", a.getSampleRoutes(), exampleMiddleware(s.Logger))
	errChan := make(chan error, 1)
	go s.Serve(errChan)

	go func() {
		serverErr := <-errChan
		if serverErr != nil {
			s.Logger.Errorf("error starting server: %v", serverErr)
			s.ShutdownServer(ctx)
		}
	}()

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigTerm
	s.Logger.Infof("received signal: %s", sig)
	s.ShutdownServer(ctx)
}
