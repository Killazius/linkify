package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"os"
	"shorturl/internal/config"
	"shorturl/internal/http-server/handlers/url/delete"
	"shorturl/internal/http-server/handlers/url/redirect"
	"shorturl/internal/http-server/handlers/url/save"
	customLogger "shorturl/internal/http-server/middleware/customLogger"
	"shorturl/internal/lib/logger"
	"shorturl/internal/lib/logger/sl"
	"shorturl/internal/storage/postgresql"
)

func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)
	storage, err := postgresql.NewStorage(cfg.StorageURL)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(customLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, storage, cfg.AliasLength))
	router.Get("/{alias}", redirect.New(log, storage))
	router.Delete("/{alias}", delete.New(log, storage))

	server := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	log.Info("starting server", "address", cfg.Address)
	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))
	}

	log.Error("server stopped")
}
