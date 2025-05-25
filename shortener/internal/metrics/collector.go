package metrics

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"linkify/internal/config"
	"net/http"
	"time"
)

type metrics struct {
	linksCreated        prometheus.Gauge
	linksRedirected     prometheus.Gauge
	linksDeleted        prometheus.Gauge
	httpRequestDuration *prometheus.HistogramVec
}
type Collector struct {
	reg *prometheus.Registry
	cfg config.Prometheus
	srv *http.Server
	log *zap.SugaredLogger
	metrics
}

func New(cfg config.Prometheus, log *zap.SugaredLogger) *Collector {
	return &Collector{
		metrics: metrics{
			linksCreated: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "url_shortener_links_created_total",
				Help: "Total number of shortened links created",
			}),
			linksRedirected: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "url_shortener_links_redirected_total",
				Help: "Total number of link redirects",
			}),
			linksDeleted: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "url_shortener_links_deleted_total",
				Help: "Total number of link redirects",
			}),
			httpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: []float64{0.1, 0.3, 0.5, 1, 3, 5},
			}, []string{"method", "path", "status"}),
		},
		reg: prometheus.NewRegistry(),
		cfg: cfg,
		log: log,
	}
}

func (c *Collector) Register() {
	c.reg.MustRegister(
		c.linksCreated,
		c.linksDeleted,
		c.httpRequestDuration,
		c.linksRedirected,
		collectors.NewGoCollector(),
	)
}
func (c *Collector) IncLinksCreated() {
	c.linksCreated.Inc()
}

func (c *Collector) IncLinksRedirected() {
	c.linksRedirected.Inc()
}

func (c *Collector) IncLinksDeleted() {
	c.linksDeleted.Inc()
}

func (c *Collector) ObserveHTTPRequestDuration(method, path, status string, duration float64) {
	c.httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)
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
		c.httpRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
			http.StatusText(rw.status),
		).Observe(duration)
	})
}
func (c *Collector) MustRun() {
	if err := c.Run(); err != nil {
		c.log.Error("failed to run collector", zap.Error(err))
	}

}

func (c *Collector) Run() error {
	mux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(c.reg, promhttp.HandlerOpts{Registry: c.reg})
	mux.Handle("/metrics", promHandler)
	c.Register()
	srv := &http.Server{
		Addr:         c.cfg.Address,
		Handler:      mux,
		ReadTimeout:  c.cfg.Timeout,
		WriteTimeout: c.cfg.Timeout,
		IdleTimeout:  c.cfg.IdleTimeout,
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
		c.log.Error("failed to stop metrics client", zap.Error(err))
	}
}
