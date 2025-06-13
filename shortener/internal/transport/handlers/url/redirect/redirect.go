package redirect

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

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=URLGetter
type URLGetter interface {
	Get(alias string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=CacheGetter
type CacheGetter interface {
	Get(ctx context.Context, key string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=MetricsGetter
type MetricsGetter interface {
	IncLinksRedirected()
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
func New(log *zap.SugaredLogger, urlGetter URLGetter, cacheGetter CacheGetter, m MetricsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			"request_id", middleware.GetReqID(r.Context()),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		url, err := cacheGetter.Get(r.Context(), alias)
		if err == nil {
			log.Infow("got url from cache", "url", url)
			m.IncLinksRedirected()
			http.Redirect(w, r, url, http.StatusFound)
			return
		}

		url, err = urlGetter.Get(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Infow("url not found", "alias", alias)
				render.Status(r, http.StatusNotFound)
				render.JSON(w, r, resp.Error("url not found"))
				return
			}
			log.Error("failed to get url", zap.Error(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to get url"))
			return
		}

		log.Infow("got url", "url", url)
		m.IncLinksRedirected()
		http.Redirect(w, r, url, http.StatusFound)
	}
}
