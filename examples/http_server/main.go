package main

import (
	"fmt"
	"log"
	"net/http"

	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
)

var routes = []cwhttp.Route{
	{
		Name:    "testing",
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

func main() {
	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(9090),
		cwhttp.SetServerApiVersion("v1"),
	)
	s.RegisterSubRouter("/yo", routes)
	log.Fatal(s.Serve())
}
