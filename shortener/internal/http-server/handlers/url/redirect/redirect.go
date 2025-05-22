package redirect

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "linkify/internal/lib/api/response"
	"linkify/internal/lib/logger/sl"
	"linkify/internal/metrics"
	"linkify/internal/storage"
	"log/slog"
	"net/http"
)

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=CacheGetter
type CacheGetter interface {
	Get(ctx context.Context, key string) (string, error)
}

// New handles the redirect of a alias by its url.
// @Summary      Redirect to URL by alias
// @Description  Redirects to the original URL using the provided alias
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        alias   path      string  true  "Alias of the URL to redirect"
// @Success      302     "Found"  "Redirects to the original URL"
// @Failure      400     {object}  response.Response  "Invalid request"
// @Failure      404     {object}  response.Response  "Alias not found"
// @Failure      500     {object}  response.Response  "Internal server error"
// @Router       /{alias} [get]
func New(log *slog.Logger, urlGetter URLGetter, cacheGetter CacheGetter, m *metrics.Collector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

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

		ctx := r.Context()

		url, err := cacheGetter.Get(ctx, alias)
		if err == nil {
			log.Info("got url from cache", slog.String("url", url))
			m.LinksRedirected.Inc()
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		url, err = urlGetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url not found", slog.String("alias", alias))
				w.WriteHeader(http.StatusNotFound)
				render.JSON(w, r, resp.Error("url not found"))
				return
			}
			log.Error("failed to get url", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get url"))
			return
		}

		log.Info("got url", slog.String("url", url))
		m.LinksRedirected.Inc()
		http.Redirect(w, r, url, http.StatusFound)
	}
}
