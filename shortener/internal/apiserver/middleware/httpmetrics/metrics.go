package httpmetrics

import (
	"linkify/internal/metrics"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func New(c *metrics.Collector) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			c.ObserveHTTPRequestDuration(r.Method, r.URL.Path, http.StatusText(rw.status), duration)
		})
	}
}
