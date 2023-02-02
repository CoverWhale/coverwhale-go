package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func newRelic(appName string) (*newrelic.Application, error) {
	_, ok := os.LookupEnv("NEW_RELIC_LICENSE_KEY")
	if !ok {
		return nil, nil
	}

	return newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigFromEnvironment(),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
}

func main() {
	ctx := context.Background()
	ds := NewInMemoryStore()
	l := logging.NewLogger()
	nr, err := newRelic("testing")
	if err != nil {
		log.Fatal(err)
	}

	h := Application{
		ProductManager: ds,
		ClientManager:  ds,
	}

	cwServer := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
		cwhttp.SetNewRelicApp(nr),
	).RegisterSubRouter("/api/v1", h.buildRoutes(l), middleware.Logger, middleware.Throttle(1))

	h.Server = cwServer

	products(ds)

	errChan := make(chan error, 1)
	go h.Server.Serve(errChan)

	go func() {
		serverErr := <-errChan
		if serverErr != nil {
			cwServer.Logger.Errorf("error starting server: %v", serverErr)
			cwServer.ShutdownServer(ctx)
		}
	}()

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigTerm
	cwServer.Logger.Infof("received signal: %s", sig)
	cwServer.ShutdownServer(ctx)

}
