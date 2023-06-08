package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CoverWhale/coverwhale-go/logging"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/ksuid"
)

func Logging(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			logging.Infof("path: %v host: %v duration: %dms", r.URL, r.Host, time.Since(start).Milliseconds())
		}()

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func RequestID(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-ID") == "" {
			id := ksuid.New()
			r.Header.Add("X-Request-ID", id.String())
			w.Header().Add("X-Request-ID", id.String())
		}

		h.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func CodeStats(h http.Handler, vec *prometheus.CounterVec, hist *prometheus.HistogramVec) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		h.ServeHTTP(ww, r)

		vec.WithLabelValues(fmt.Sprintf("%d", ww.Status()), r.Method, r.URL.Path).Inc()
		hist.WithLabelValues(fmt.Sprintf("%d", ww.Status()), r.Method, r.URL.Path).Observe(float64(time.Since(start).Seconds()))

	}

	return http.HandlerFunc(fn)

}
