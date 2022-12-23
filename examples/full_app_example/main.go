package main

import (
	"log"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
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

	log.Fatal(h.Server.Serve())

}
