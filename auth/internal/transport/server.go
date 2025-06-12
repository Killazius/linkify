package transport

import (
	"auth/internal/config"
	"auth/internal/service"
	"auth/internal/transport/handlers"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	server *http.Server
	router *chi.Mux
	log    *zap.SugaredLogger
	repo   service.Repository
	config config.HTTPConfig
}

func New(
	cfg config.HTTPConfig,
	repo service.Repository,
	log *zap.SugaredLogger,
) *Server {
	router := chi.NewRouter()
	srv := &Server{
		server: &http.Server{
			Addr:         "0.0.0.0:" + cfg.Port,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
			Handler:      router,
		},
		router: router,
		log:    log,
		repo:   repo,
		config: cfg,
	}

	srv.registerRoutes()
	return srv
}

func (s *Server) registerRoutes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.URLFormat)
	s.router.Post("/register", handlers.Register(s.log, s.repo))
	s.router.Post("/login", handlers.Login(s.log, s.repo))
	s.router.Post("/refresh", handlers.Refresh(s.log, s.repo))

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
	//err := s.repo.Stop()
	//if err != nil {
	//	s.log.Error("failed to stop repository client", zap.Error(err))
	//}
}
