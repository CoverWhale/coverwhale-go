package main

import (
	"fmt"
	"log"
	"net/http"

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
	if r.Header.Get("Authorization") == "" {
		return cwhttp.NewAppError(fmt.Errorf("unauthorized"), 401)
	}

	if ie != "" {
		return fmt.Errorf("this is an internal error")
	}

	if ce != "" {
		return cwhttp.NewAppError(fmt.Errorf("uh oh something is wrong"), 400)
	}

	w.Write([]byte("this works!"))
	return nil
}

func exampleMiddleware(s *cwhttp.Server) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s.Logger.Info("in middleware")
			h.ServeHTTP(w, r)
		})
	}

}

func main() {
	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
	)

	s.RegisterSubRouter("/api/v1", routes, exampleMiddleware(s))
	log.Fatal(s.Serve())
}
