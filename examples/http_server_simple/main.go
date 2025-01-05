// Copyright 2025 Sencillo
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
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/SencilloDev/sencillo-go/metrics"
	cwhttp "github.com/SencilloDev/sencillo-go/transports/http"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

func getRoutes(l *slog.Logger) []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method: http.MethodGet,
			Path:   "/testing-things/{testID}",
			Handler: &cwhttp.ErrHandler{
				Handler: testing,
				Logger:  l,
			},
		},
	}
}

func doMore(ctx context.Context) {
	// create new span from context
	_, span := metrics.NewTracer(ctx, "more sleepy")
	defer span.End()

	time.Sleep(500 * time.Millisecond)
}

func testing(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("testID")
	ie := r.Header.Get("internal-error")
	ce := r.Header.Get("client-error")

	slog.Info(id)

	if ie != "" {
		return fmt.Errorf("this is an internal error")
	}

	if ce != "" {
		return cwhttp.NewClientError(fmt.Errorf("uh oh something is wrong"), 400)
	}

	// get new span
	ctx, span := metrics.NewTracer(r.Context(), "sleepytime")

	// if wanted define attributes for span
	attrs := []attribute.KeyValue{
		attribute.String("test", "this"),
	}
	span.SetAttributes(attrs...)
	defer span.End()

	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(400-90+1) + 90

	sleep := time.Duration(i) * time.Millisecond
	time.Sleep(sleep)

	// fake call to something that takes a long time
	doMore(ctx)

	resp := fmt.Sprintf("this works and took %dms\n", sleep.Milliseconds())

	w.Write([]byte(resp))
	return nil
}

func exampleMiddleware(l *slog.Logger) func(h http.Handler) http.Handler {
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

	// create new metrics exporter
	exp, err := metrics.NewOTLPExporter(ctx, "localhost:4318", otlptracehttp.WithInsecure())
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// create global tracer provider
	tp, err := metrics.RegisterGlobalOTLPProvider(exp, "simple-http-example", "v1")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(7070),
		cwhttp.SetTracerProvider(tp),
	)

	s.RegisterSubRouter("/api/v1/", getRoutes(s.Logger), exampleMiddleware(s.Logger))

	errChan := make(chan error, 1)
	go s.Serve(errChan)
	s.AutoHandleErrors(ctx, errChan)
}
