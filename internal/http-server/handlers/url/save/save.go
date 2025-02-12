package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	resp "shorturl/internal/lib/api/response"
	"shorturl/internal/lib/logger/sl"
	"shorturl/internal/lib/random"
	"shorturl/internal/storage"
	"time"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias     string    `json:"alias"`
	CreatedAt time.Time `json:"created_at"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string, createdAt time.Time) error
}

func New(log *slog.Logger, urlSaver URLSaver, aliasLength int) http.HandlerFunc {
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

			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErrs validator.ValidationErrors
			errors.As(err, &validateErrs)

			log.Error("failed to validate request", sl.Err(err))

			render.JSON(w, r, resp.ValidateError(validateErrs))
			return
		}
		alias := req.Alias

		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		err = urlSaver.SaveURL(req.URL, alias, time.Now())
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				log.Info("url already exists", "url", req.URL)

				render.JSON(w, r, resp.Error("url already exists"))
				return
			}
			log.Error("failed to save url", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to save url"))
			return
		}

		log.Info("new URL added", "url", req.URL)

		responseOK(w, r, alias, time.Now())
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string, createdAt time.Time) {
	render.JSON(w, r, Response{
		Response:  resp.OK(),
		Alias:     alias,
		CreatedAt: createdAt,
	})
}
