package main

import (
	"log"
	"net/http"

	"github.com/CoverWhale/coverwhale-go/logging"
	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
)

type DataStore interface {
	GetData(id string) string
}

type Server interface {
	Serve() error
}

type App struct {
	Server Server
	DS     DataStore
}

type InMem struct {
	Data map[string]string
}

func (i InMem) GetData(id string) string {
	return i.Data[id]
}

func (a *App) getSampleRoutes() []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method:  http.MethodGet,
			Path:    "/testing",
			Handler: http.HandlerFunc(a.exampleGet),
		},
		{
			Method:  http.MethodGet,
			Path:    "/testingCustom",
			Handler: cwhttp.HandleWithContext(customHandlerType, a.DS),
		},
	}
}

func (a *App) exampleGet(w http.ResponseWriter, r *http.Request) {

	id := r.URL.Query().Get("id")

	data := a.DS.GetData(id)

	w.Write([]byte(data))
}

func customHandlerType(w http.ResponseWriter, r *http.Request, ds DataStore) {
	id := r.URL.Query().Get("id")

	data := ds.GetData(id)

	w.Write([]byte(data))
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
	ds := InMem{
		Data: map[string]string{
			"test": "testing",
		},
	}

	a := App{
		Server: s,
		DS:     ds,
	}

	s.RegisterSubRouter("/api/v1", a.getSampleRoutes(), exampleMiddleware(s.Logger))
	log.Fatal(s.Serve())
}
