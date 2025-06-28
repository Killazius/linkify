package transport

import (
	"context"
	"errors"
	"fmt"
	"github.com/Killazius/linkify-proto/pkg/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"linkify/internal/config"
	"linkify/internal/metrics"
	"linkify/internal/transport/handlers/url/delete"
	"linkify/internal/transport/handlers/url/redirect"
	"linkify/internal/transport/handlers/url/save"
	"linkify/internal/transport/middleware/auth"
	customLogger "linkify/internal/transport/middleware/customLogger"
	"linkify/internal/transport/middleware/httpmetrics"
	"net/http"
	"time"
)

type Repository interface {
	Save(urlToSave string, alias string, createdAt time.Time) error
	Get(alias string) (string, error)
	Delete(alias string) error
	Stop() error
}

type Cache interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Stop() error
}
type Auth interface {
	ValidateToken(ctx context.Context, in *api.TokenRequest, opts ...grpc.CallOption) (*api.TokenResponse, error)
}

type Server struct {
	server  *http.Server
	router  *chi.Mux
	log     *zap.SugaredLogger
	repo    Repository
	cache   Cache
	metrics *metrics.Collector
	config  config.HTTPServer
	client  Auth
}

func New(
	cfg config.HTTPServer,
	log *zap.SugaredLogger,
	repo Repository,
	cache Cache,
	metrics *metrics.Collector,
	client Auth,

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
		repo:    repo,
		cache:   cache,
		metrics: metrics,
		config:  cfg,
		client:  client,
	}

	srv.registerRoutes()
	return srv
}

func (s *Server) registerRoutes() {
	s.router.Use(httpmetrics.New(s.metrics))
	s.router.Use(middleware.RequestID)
	s.router.Use(customLogger.New(s.log))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.URLFormat)
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://127.0.0.1"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	s.router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", s.config.IP)),
	))
	s.router.Get("/{alias}", redirect.New(s.log, s.repo, s.cache, s.metrics))

	s.router.With(auth.New(s.client, s.log)).Route("/api", func(r chi.Router) {
		r.Post("/url", save.New(s.log, s.repo, s.cache, s.config.AliasLength, s.metrics))
		r.Delete("/url/{alias}", delete.New(s.log, s.repo, s.cache, s.metrics))
	})
}
func (s *Server) MustRun() {
	if err := s.Run(); err != nil {
		s.log.Fatal("failed to run HTTP-server", zap.Error(err))
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
		s.log.Error("failed to stop HTTP server", zap.Error(err))
	}
	err := s.cache.Stop()
	if err != nil {
		s.log.Error("failed to stop redis client", zap.Error(err))
	}
	err = s.repo.Stop()
	if err != nil {
		s.log.Error("failed to stop repository client", zap.Error(err))
	}
}
