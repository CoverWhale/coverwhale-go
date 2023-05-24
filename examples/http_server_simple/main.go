package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
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

	rand.Seed(time.Now().UnixNano())

	i := rand.Intn(400-90+1) + 90
	sleep := time.Duration(i) * time.Millisecond
	time.Sleep(sleep)

	resp := fmt.Sprintf("this works and took %dms\n", sleep.Milliseconds())

	w.Write([]byte(resp))
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

	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(7070),
	)

	s.RegisterSubRouter("/api/v1", getRoutes(s.Logger), exampleMiddleware(s.Logger))

	errChan := make(chan error, 1)
	go s.Serve(errChan)
	s.AutoHandleErrors(ctx, errChan)
}
