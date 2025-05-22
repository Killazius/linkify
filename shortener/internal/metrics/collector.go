package metrics

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"time"
)

type metrics struct {
	LinksCreated        prometheus.Gauge
	LinksRedirected     prometheus.Gauge
	LinksDeleted        prometheus.Gauge
	HTTPRequestDuration *prometheus.HistogramVec
}
type Collector struct {
	reg     *prometheus.Registry
	address string
	srv     *http.Server
	metrics
}

func New(address string) *Collector {
	return &Collector{
		metrics: metrics{
			LinksCreated: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "url_shortener_links_created_total",
				Help: "Total number of shortened links created",
			}),
			LinksRedirected: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "url_shortener_links_redirected_total",
				Help: "Total number of link redirects",
			}),
			LinksDeleted: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "url_shortener_links_deleted_total",
				Help: "Total number of link redirects",
			}),
			HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: []float64{0.1, 0.3, 0.5, 1, 3, 5},
			}, []string{"method", "path", "status"}),
		},
		reg:     prometheus.NewRegistry(),
		address: address,
	}
}

func (c *Collector) Register() {
	c.reg.MustRegister(
		c.LinksCreated,
		c.LinksDeleted,
		c.HTTPRequestDuration,
		c.LinksRedirected,
		collectors.NewGoCollector(),
	)
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (c *Collector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		c.HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			http.StatusText(rw.status),
		).Observe(duration)
	})
}
func (c *Collector) MustRun(log *slog.Logger) {
	if err := c.Run(); err != nil {
		log.Error(err.Error())
	}

}

func (c *Collector) Run() error {
	mux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(c.reg, promhttp.HandlerOpts{Registry: c.reg})
	mux.Handle("/metrics", promHandler)

	srv := &http.Server{
		Addr:    c.address,
		Handler: mux,
	}
	c.srv = srv
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (c *Collector) Stop(ctx context.Context) {
	if err := c.srv.Shutdown(ctx); err != nil {
	}
}
