package deleteAlias

import (
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

func New(log *slog.Logger, URLDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.deleteAlias.New"

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

		log.Info("delete alias", slog.String("alias", alias))
		w.WriteHeader(http.StatusNoContent)
	}
}
