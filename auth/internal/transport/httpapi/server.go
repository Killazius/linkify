package httpapi

import (
	"auth/internal/config"
	"auth/internal/transport/grpcapi"
	"auth/internal/transport/httpapi/handlers"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	server *http.Server
	log    *zap.SugaredLogger
}

func NewServer(
	log *zap.SugaredLogger,
	authService grpcapi.Repository,
	cfg config.HTTPConfig,
) *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	authHandler := handlers.NewAuthHandler(log, authService)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register())
		r.Post("/login", authHandler.Login())
		r.Post("/refresh", authHandler.Refresh())
		r.Delete("/logout", authHandler.Logout())
		r.Delete("/account", authHandler.DeleteAccount())
	})

	return &Server{
		log: log,
		server: &http.Server{
			Addr:         "0.0.0.0:" + cfg.Port,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
			Handler:      r,
		},
	}
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
}
