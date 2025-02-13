package delete

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "shorturl/internal/lib/api/response"
	"shorturl/internal/lib/logger/sl"
	"shorturl/internal/storage"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

type CacheDeleter interface {
	Delete(ctx context.Context, key string) error
}

func New(log *slog.Logger, URLDeleter URLDeleter, CacheDeleter CacheDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}
		err := URLDeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("alias not found")
				render.JSON(w, r, resp.Error("alias not found"))
				return
			}
			log.Error("failed to get alias", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to get alias"))
			return
		}
		ctx := r.Context()
		err = CacheDeleter.Delete(ctx, alias)
		if err != nil {
			log.Error("failed to delete alias from cache", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to delete alias from cache"))
			return
		}
		log.Info("delete alias", slog.String("alias", alias))
		w.WriteHeader(http.StatusNoContent)
	}
}
