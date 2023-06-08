package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()
	ds := NewInMemoryStore()
	l := logging.NewLogger()

	h := Application{
		ProductManager: ds,
		ClientManager:  ds,
	}

	cwServer := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
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
