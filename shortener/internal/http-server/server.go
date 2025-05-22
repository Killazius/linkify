package http_server

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"linkify/internal/config"
	"linkify/internal/http-server/handlers/url/delete"
	"linkify/internal/http-server/handlers/url/redirect"
	"linkify/internal/http-server/handlers/url/save"
	customLogger "linkify/internal/http-server/middleware/customLogger"
	"linkify/internal/lib/logger/sl"
	"linkify/internal/metrics"
	"linkify/internal/storage/cache"
	"linkify/internal/storage/postgresql"
	"log/slog"
	"net/http"
)

type Server struct {
	server  *http.Server
	router  *chi.Mux
	log     *slog.Logger
	storage *postgresql.Storage
	redis   *cache.Storage
	metrics *metrics.Collector
	config  config.HTTPServer
}

func New(
	cfg config.HTTPServer,
	log *slog.Logger,
	storage *postgresql.Storage,
	redis *cache.Storage,
	metrics *metrics.Collector,
) *Server {
	router := chi.NewRouter()
	srv := &Server{
		server: &http.Server{
			Addr:         cfg.Address,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
			Handler:      router,
		},
		router:  router,
		log:     log,
		storage: storage,
		redis:   redis,
		metrics: metrics,
		config:  cfg,
	}

	srv.registerRoutes()
	return srv
}

func (s *Server) registerRoutes() {
	s.router.Use(s.metrics.Middleware)
	s.router.Use(middleware.RequestID)
	s.router.Use(customLogger.New(s.log))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.URLFormat)

	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", s.config.IP)),
	))

	s.router.Post("/url", save.New(s.log, s.storage, s.redis, s.config.AliasLength, s.metrics))
	s.router.Get("/{alias}", redirect.New(s.log, s.storage, s.redis, s.metrics))
	s.router.Delete("/url/{alias}", delete.New(s.log, s.storage, s.redis, s.metrics))
}

func (s *Server) MustRun() {
	if err := s.Run(); err != nil {
		s.log.Error(err.Error())
	}
}
func (s *Server) Run() error {
	err := s.server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
func (s *Server) Stop(ctx context.Context) {
	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Error("failed to stop HTTP server", sl.Err(err))
	}
	err := s.redis.Stop()
	if err != nil {
		s.log.Error("failed to stop redis client", sl.Err(err))
	}
	err = s.storage.Stop()
	if err != nil {
		s.log.Error("failed to stop storage client", sl.Err(err))
	}
}
