package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	cwhttp "github.com/CoverWhale/coverwhale-go/transports/http"
	"github.com/CoverWhale/coverwhale-go/transports/http/middleware"
	"github.com/CoverWhale/logr"
)

type Request struct {
	SampleValue string `json:"sample_value"`
}

func getRoutes(l *logr.Logger) []cwhttp.Route {
	return []cwhttp.Route{
		{
			Method:  http.MethodGet,
			Path:    "/test",
			Handler: http.HandlerFunc(test),
		},
		{
			Method:  http.MethodPost,
			Path:    "/test-custom",
			Handler: middleware.CustomValidator(http.HandlerFunc(test), "http://localhost:8181", "cw/underwriting"),
		},
	}
}

func test(w http.ResponseWriter, r *http.Request) {
	var req Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logr.Errorf("error decoding app data: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := fmt.Sprintf("%s", req.SampleValue)

	w.Write([]byte(resp))
}

func main() {
	ctx := context.Background()

	s := cwhttp.NewHTTPServer(
		cwhttp.SetServerPort(7070),
	)

	s.RegisterSubRouter("/api/v1", getRoutes(s.Logger), middleware.RequestID)

	errChan := make(chan error, 1)
	go s.Serve(errChan)
	s.AutoHandleErrors(ctx, errChan)
}
