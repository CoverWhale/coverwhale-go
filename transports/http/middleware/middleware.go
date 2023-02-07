package middleware

import (
	"net/http"
	"time"

	"github.com/CoverWhale/coverwhale-go/logging"
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
