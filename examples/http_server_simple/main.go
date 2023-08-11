package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/CoverWhale/coverwhale-go/metrics"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/CoverWhale/logr"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

func getRoutes(l *logr.Logger) []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method: http.MethodGet,
			Path:   "/testing/{testID}",
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
	id := chi.URLParam(r, "testID")
	ie := r.Header.Get("internal-error")
	ce := r.Header.Get("client-error")

	fmt.Println(id)

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

	// fake call to somethign that takes a long time
	doMore(ctx)

	resp := fmt.Sprintf("this works and took %dms\n", sleep.Milliseconds())

	w.Write([]byte(resp))
	return nil
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

	// create new metrics exporter
	exp, err := metrics.NewOTLPExporter(ctx, "localhost:4318", otlptracehttp.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	// create global tracer provider
	tp, err := metrics.RegisterGlobalOTLPProvider(exp, "simple-http-example", "v1")
	if err != nil {
		log.Fatal(err)
	}

	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(7070),
		cwhttp.SetTracerProvider(tp),
	)

	s.RegisterSubRouter("/api/v1", getRoutes(s.Logger), exampleMiddleware(s.Logger))

	errChan := make(chan error, 1)
	go s.Serve(errChan)
	s.AutoHandleErrors(ctx, errChan)
}
