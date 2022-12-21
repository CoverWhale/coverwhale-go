package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
)

var routes = []cwhttp.Route{
	{
		Method:  http.MethodGet,
		Path:    "/testing",
		Handler: testing,
	},
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
	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
	)

	s.RegisterSubRouter("/api/v1", routes, exampleMiddleware(s.Logger))
	log.Fatal(s.Serve())
}
