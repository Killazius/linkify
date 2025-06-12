package handlers

import (
	"auth/internal/lib/jwt"
	"auth/internal/repository"
	"auth/internal/transport/grpcapi"
	"errors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type AuthHandler struct {
	log  *zap.SugaredLogger
	repo grpcapi.Repository
}

func NewAuthHandler(log *zap.SugaredLogger, authService grpcapi.Repository) *AuthHandler {
	return &AuthHandler{
		log:  log,
		repo: authService,
	}
}
func (h *AuthHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			h.log.Error("failed to decode request", zap.Error(err))

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, "implement me")
			return
		}
		uid, err := h.repo.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, repository.ErrInvalidCredentials) {
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, "implement me")
				return
			}
			h.log.Error("failed to register user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, "implement me")
			return
		}
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, uid)
	}
}

func (h *AuthHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			h.log.Error("failed to decode login request", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid request format"})
			return
		}

		accessToken, refreshToken, err := h.repo.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, repository.ErrInvalidCredentials) {
				h.log.Warn("invalid login attempt", zap.String("email", req.Email))
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, map[string]string{"error": "invalid credentials"})
				return
			}

			h.log.Error("failed to login user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}

		repoImpl, ok := h.repo.(*repository.Repository)
		if !ok {
			h.log.Error("invalid repository type")
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

func (h *AuthHandler) Refresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			h.log.Warn("refresh token cookie not found", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "refresh token required"})
			return
		}

		newAccessToken, newRefreshToken, err := h.repo.RefreshTokens(r.Context(), cookie.Value)
		if err != nil {
			h.log.Error("failed to refresh tokens", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}

		repoImpl, ok := h.repo.(*repository.Repository)
		if !ok {
			h.log.Error("invalid repository type")
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
func (h *AuthHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := getRefreshToken(r)
		if err != nil || t == "" {
			h.log.Warn("refresh token cookie not found", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "invalid refresh token"})
			return
		}
		err = h.repo.Logout(r.Context(), t)
		if err != nil {
			h.log.Error("failed to logout", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}
		clearAuthCookies(w)
		w.WriteHeader(http.StatusOK)
	}
}
func (h *AuthHandler) DeleteAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := getRefreshToken(r)
		if err != nil {
			h.log.Warn("refresh token cookie not found", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "invalid refresh token"})
			return
		}
		user, err := jwt.VerifyToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, _ := strconv.ParseInt(user.ID, 10, 64)
		if err = h.repo.DeleteAccount(r.Context(), userID); err != nil {
			h.log.Error("failed to delete account", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		clearAuthCookies(w)

		w.WriteHeader(http.StatusNoContent)
	}
}

func getRefreshToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
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
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(refreshTTL.Seconds()),
	})
}
func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}
