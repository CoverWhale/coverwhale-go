package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func getRoutes(l *logging.Logger) []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method: http.MethodGet,
			Path:   "/testing",
			Handler: &cwhttp.ErrHandler{
				Handler: testing,
				Logger:  l,
			},
		},
	}
}

func testing(w http.ResponseWriter, r *http.Request) error {
	ie := r.Header.Get("internal-error")
	ce := r.Header.Get("client-error")

	if ie != "" {
		return fmt.Errorf("this is an internal error")
	}

	if ce != "" {
		return cwhttp.NewClientError(fmt.Errorf("uh oh something is wrong"), 400)
	}

	w.Write([]byte("this works!"))
	return nil
}

func exampleMiddleware(l *logging.Logger) func(h http.Handler) http.Handler {
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
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("testing"),
		newrelic.ConfigFromEnvironment(),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
		cwhttp.SetNewRelicApp(app),
	)

	s.RegisterSubRouter("/api/v1", getRoutes(s.Logger), exampleMiddleware(s.Logger))

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
