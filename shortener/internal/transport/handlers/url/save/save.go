package save

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"linkify/internal/lib/api/response"
	"linkify/internal/lib/random"
	"linkify/internal/storage"

	"github.com/go-chi/render"
	"net/http"
	"time"
)

type Request struct {
	URL string `json:"url" validate:"required,url"`
}

// Response represents the response structure for the save handler.
// @Description Response contains the status, alias, and creation time of the saved URL.
type Response struct {
	// Status is the response status.
	// @Example success
	// @Example error
	// @json:inline
	response.Response `swaggertype:"object,string"`

	Alias string `json:"alias"`

	CreatedAt time.Time `json:"created_at"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=URLSaver
type URLSaver interface {
	Save(urlToSave string, alias string, createdAt time.Time) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=CacheSaver
type CacheSaver interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=MetricsSaver
type MetricsSaver interface {
	IncLinksCreated()
}

// New handles the save of a URL by its alias.
// @Summary      Save URL
// @Description  Saves a URL and generates a unique alias
// @Tags         url
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body Request true "URL to save"
// @Success      201  {object}  Response  "URL saved successfully"
// @Failure      400  {object}  response.Response  "Invalid request or validation error"
// @Failure      401  {object}  response.Response  "Unauthorized"
// @Failure      500  {object}  response.Response  "Internal server error"
// @Router       /api/url [post]
func New(log *zap.SugaredLogger, urlSaver URLSaver, CacheSaver CacheSaver, aliasLength int, m MetricsSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(
			"request_id", middleware.GetReqID(r.Context()),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", zap.Error(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}

		log.Infow("request body decoded", "request", req)

		if err = validator.New().Struct(req); err != nil {
			var validateErrs validator.ValidationErrors
			errors.As(err, &validateErrs)

			log.Error("failed to validate request", zap.Error(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidateError(validateErrs))
			return
		}
		now := time.Now()
		alias, err := generateUniqueAlias(log, urlSaver, req.URL, aliasLength, now)
		if err != nil {
			log.Error("failed to generate unique alias after multiple attempts", zap.Error(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to generate unique alias"))
			return
		}
		if err := CacheSaver.Set(r.Context(), alias, req.URL, time.Hour); err != nil {
			log.Errorw("failed to save in cache", "alias", alias, "error", err)
		}
		m.IncLinksCreated()
		log.Infow("new URL added", "url", req.URL)
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{
			Response:  response.OK(),
			Alias:     alias,
			CreatedAt: now,
		})
	}
}

func generateUniqueAlias(log *zap.SugaredLogger, saver URLSaver, url string, length int, createdAt time.Time) (string, error) {
	const maxAttempts = 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		alias := random.NewRandomString(length)
		err := saver.Save(url, alias, createdAt)
		if err == nil {
			return alias, nil
		}

		if !errors.Is(err, storage.ErrAliasExists) {
			return "", err
		}

		log.Infow("alias collision", "attempt", attempt+1, "alias", alias)
	}

	return "", errors.New("max attempts reached")
}
