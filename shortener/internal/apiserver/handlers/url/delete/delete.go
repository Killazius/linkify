package delete

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	resp "linkify/internal/lib/api/response"
	"linkify/internal/storage"
	"net/http"
)

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=URLDeleter
type URLDeleter interface {
	Delete(alias string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=CacheDeleter
type CacheDeleter interface {
	Delete(ctx context.Context, key string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=MetricsDeleter
type MetricsDeleter interface {
	IncLinksDeleted()
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
func New(log *zap.SugaredLogger, URLDeleter URLDeleter, CacheDeleter CacheDeleter, m MetricsDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			"request_id", middleware.GetReqID(r.Context()),
		)
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}
		err := CacheDeleter.Delete(r.Context(), alias)
		if err != nil {
			log.Error("failed to delete alias from cache", zap.Error(err))
		}
		err = URLDeleter.Delete(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Infow("alias not found", "alias", alias)
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.Error("alias not found"))
				return
			}
			log.Error("failed to get alias", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get alias"))
			return
		}
		log.Infow("delete alias", "alias", alias)
		m.IncLinksDeleted()
		w.WriteHeader(http.StatusNoContent)
	}
}
