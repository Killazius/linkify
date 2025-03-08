package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "go.uber.org/automaxprocs"
	"net/http"
	"os"
	_ "shorturl/docs"
	"shorturl/internal/config"
	"shorturl/internal/http-server/handlers/url/delete"
	"shorturl/internal/http-server/handlers/url/redirect"
	"shorturl/internal/http-server/handlers/url/save"
	customLogger "shorturl/internal/http-server/middleware/customLogger"
	"shorturl/internal/lib/logger"
	"shorturl/internal/lib/logger/sl"
	"shorturl/internal/storage/cache"
	"shorturl/internal/storage/postgresql"
)

// @title           Linkify
// @version         1.2
// @description     Link shortening service.
// @host            localhost:8080

// @contact.name Telegram Developer
// @contact.url https://t.me/killazDev

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)
	storage, err := postgresql.NewStorage(cfg.StorageURL)
	if err != nil {
		log.Error("failed to initialize storage", sl.Err(err))
		os.Exit(1)
	}
	redis, err := cache.NewStorage(cfg.Addr, cfg.Password, cfg.DB)
	if err != nil {
		log.Error("failed to initialize cache", sl.Err(err))
		os.Exit(1)
	}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(customLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))
	router.Post("/url", save.New(log, storage, redis, cfg.AliasLength))
	router.Get("/{alias}", redirect.New(log, storage, redis))
	router.Delete("/{alias}", delete.New(log, storage, redis))
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
