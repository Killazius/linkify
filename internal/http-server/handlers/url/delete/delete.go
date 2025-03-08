package delete

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "linkify/internal/lib/api/response"
	"linkify/internal/lib/logger/sl"
	"linkify/internal/storage"
	"log/slog"
	"net/http"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

type CacheDeleter interface {
	Delete(ctx context.Context, key string) error
}

// New handles the deletion of a URL by its alias.
// @Summary      Delete alias for URL
// @Description  Delete URL by alias
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        alias   path      string  true  "Alias of the URL to delete"
// @Success      204     "No Content"
// @Failure      400     {object}  response.Response  "Invalid request"
// @Failure      404     {object}  response.Response  "Alias not found"
// @Failure      500     {object}  response.Response  "Internal server error"
// @Router       /{alias} [delete]
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

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}
		err := URLDeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("alias not found")

				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.Error("alias not found"))
				return
			}
			log.Error("failed to get alias", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get alias"))
			return
		}
		ctx := r.Context()
		err = CacheDeleter.Delete(ctx, alias)
		if err != nil {
			log.Error("failed to delete alias from cache", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to delete alias from cache"))
			return
		}
		log.Info("delete alias", slog.String("alias", alias))
		w.WriteHeader(http.StatusNoContent)
	}
}
