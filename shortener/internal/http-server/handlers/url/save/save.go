package save

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"linkify/internal/lib/api/response"
	"linkify/internal/lib/logger/sl"
	"linkify/internal/lib/random"
	"linkify/internal/storage"

	"github.com/go-chi/render"
	"log/slog"
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
	SaveURL(urlToSave string, alias string, createdAt time.Time) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=CacheSaver
type CacheSaver interface {
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
}

// New handles the save of a URL by its alias.
// @Summary      Save URL for alias
// @Description  Save alias by URL
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        request body Request true "Request body"
// @Success      200  {object}  Response  "URL saved successfully"
// @Failure      400  {object}  response.Response  "Invalid request"
// @Failure      500  {object}  response.Response  "Internal server error"
// @Router       /url [post]
func New(log *slog.Logger, urlSaver URLSaver, CacheSaver CacheSaver, aliasLength int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErrs validator.ValidationErrors
			errors.As(err, &validateErrs)

			log.Error("failed to validate request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, response.ValidateError(validateErrs))
			return
		}
		now := time.Now()
		maxAttempts := 5
		var alias string
		for attempt := 0; attempt < maxAttempts; attempt++ {
			alias = random.NewRandomString(aliasLength)
			err = urlSaver.SaveURL(req.URL, alias, now)
			if err == nil {
				break
			}

			if !errors.Is(err, storage.ErrAliasExists) {
				log.Error("failed to save url", sl.Err(err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, response.Error("failed to save url"))
				return
			}

			log.Info("alias already exists, generating new one", "alias", alias)
		}

		if err != nil {
			log.Error("failed to generate unique alias after multiple attempts", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to generate unique alias"))
			return
		}
		ctx := r.Context()
		err = CacheSaver.Set(ctx, alias, req.URL, 1*time.Hour)
		if err != nil {
			if errors.Is(err, storage.ErrAliasExists) {
				log.Info("alias already exists", "alias", alias)
				render.JSON(w, r, response.Error("alias already exists"))
				return
			}
			render.JSON(w, r, response.Error("failed to save alias in cache"))
			return
		}
		log.Info("url saved in cache", "alias", alias, "url", req.URL)

		log.Info("new URL added", "url", req.URL)

		responseOK(w, r, alias, time.Now())
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string, createdAt time.Time) {
	render.JSON(w, r, Response{
		Response:  response.OK(),
		Alias:     alias,
		CreatedAt: createdAt,
	})
}
