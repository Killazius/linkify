package handlers

import (
	"auth/internal/repository"
	"auth/internal/service"
	"errors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func Register(log *zap.SugaredLogger, repo service.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", zap.Error(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "implement me")
			return
		}
		uid, err := repo.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, repository.ErrInvalidCredentials) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, "implement me")
				return
				//return nil, status.Error(codes.AlreadyExists, "user already exists")
			}
			log.Error("failed to register user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, "implement me")
			return
		}
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, uid)
	}
}

func Login(log *zap.SugaredLogger, repo service.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode login request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid request format"})
			return
		}

		accessToken, refreshToken, err := repo.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, repository.ErrInvalidCredentials) {
				log.Warn("invalid login attempt", zap.String("email", req.Email))
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "invalid credentials"})
				return
			}

			log.Error("failed to login user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}

		repoImpl, ok := repo.(*repository.Repository)
		if !ok {
			log.Error("invalid repository type")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}

		setAuthCookies(w, accessToken, refreshToken, repoImpl.AccessTokenTTL, repoImpl.RefreshTokenTTL)
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"access_token_expires_in":  int(repoImpl.AccessTokenTTL.Seconds()),
			"refresh_token_expires_in": int(repoImpl.RefreshTokenTTL.Seconds()),
		})
	}
}

func Refresh(log *zap.SugaredLogger, repo service.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			log.Warn("refresh token cookie not found", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "refresh token required"})
			return
		}

		newAccessToken, newRefreshToken, err := repo.RefreshTokens(r.Context(), cookie.Value)
		if err != nil {
			log.Error("failed to refresh tokens", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}

		repoImpl, ok := repo.(*repository.Repository)
		if !ok {
			log.Error("invalid repository type")
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}

		setAuthCookies(w, newAccessToken, newRefreshToken, repoImpl.AccessTokenTTL, repoImpl.RefreshTokenTTL)

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"access_token_expires_in":  int(repoImpl.AccessTokenTTL.Seconds()),
			"refresh_token_expires_in": int(repoImpl.RefreshTokenTTL.Seconds()),
		})
	}
}
func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(accessTTL.Seconds()),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/refresh",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(refreshTTL.Seconds()),
	})
}
