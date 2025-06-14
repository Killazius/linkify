package handlers

import (
	"auth/internal/lib/jwt"
	"auth/internal/repository"
	"auth/internal/transport"
	"errors"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type AuthHandler struct {
	log  *zap.SugaredLogger
	repo transport.Repository
}
type ErrorResponse struct {
	Error string `json:"error"`
}
type TokenResponse struct {
	AccessTokenExpiresIn  int `json:"access_token_expires_in"`
	RefreshTokenExpiresIn int `json:"refresh_token_expires_in"`
}

func NewAuthHandler(log *zap.SugaredLogger, authService transport.Repository) *AuthHandler {
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

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			h.log.Error("failed to decode request", zap.Error(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrorResponse{"invalid request format"})
			return
		}

		uid, err := h.repo.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrInvalidCredentials):
				h.log.Warn("registration failed - user exists", zap.String("email", req.Email))
				render.Status(r, http.StatusConflict)
				render.JSON(w, r, ErrorResponse{"user already exists"})
			default:
				h.log.Error("failed to register user", zap.Error(err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, ErrorResponse{"internal server error"})
			}
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, map[string]int64{"user_id": uid})
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
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrorResponse{"invalid request format"})
			return
		}

		accessToken, refreshToken, err := h.repo.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrInvalidCredentials):
				h.log.Warn("invalid login attempt", zap.String("email", req.Email))
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, ErrorResponse{"invalid credentials"})
			default:
				h.log.Error("failed to login user", zap.Error(err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, ErrorResponse{"internal server error"})
			}
			return
		}

		repoImpl, ok := h.repo.(*repository.Repository)
		if !ok {
			h.log.Error("invalid repository type")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, ErrorResponse{"internal server error"})
			return
		}

		setAuthCookies(w, accessToken, refreshToken, repoImpl.AccessTokenTTL, repoImpl.RefreshTokenTTL)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, TokenResponse{
			AccessTokenExpiresIn:  int(repoImpl.AccessTokenTTL.Seconds()),
			RefreshTokenExpiresIn: int(repoImpl.RefreshTokenTTL.Seconds()),
		})
	}
}

func (h *AuthHandler) Refresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			h.log.Warn("refresh token cookie not found", zap.Error(err))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, ErrorResponse{"refresh token required"})
			return
		}

		newAccessToken, newRefreshToken, err := h.repo.RefreshTokens(r.Context(), cookie.Value)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrInvalidCredentials):
				h.log.Warn("invalid refresh token")
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, ErrorResponse{"invalid refresh token"})
			default:
				h.log.Error("failed to refresh tokens", zap.Error(err))
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, ErrorResponse{"internal server error"})
			}
			return
		}

		repoImpl, ok := h.repo.(*repository.Repository)
		if !ok {
			h.log.Error("invalid repository type")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, ErrorResponse{"internal server error"})
			return
		}

		setAuthCookies(w, newAccessToken, newRefreshToken, repoImpl.AccessTokenTTL, repoImpl.RefreshTokenTTL)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, TokenResponse{
			AccessTokenExpiresIn:  int(repoImpl.AccessTokenTTL.Seconds()),
			RefreshTokenExpiresIn: int(repoImpl.RefreshTokenTTL.Seconds()),
		})
	}
}

func (h *AuthHandler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := getRefreshToken(r)
		if err != nil {
			h.log.Warn("refresh token cookie not found", zap.Error(err))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, ErrorResponse{"refresh token required"})
			return
		}

		if err := h.repo.Logout(r.Context(), t); err != nil {
			h.log.Error("failed to logout", zap.Error(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, ErrorResponse{"internal server error"})
			return
		}

		clearAuthCookies(w)
		render.Status(r, http.StatusNoContent)
		render.NoContent(w, r)
	}
}

func (h *AuthHandler) DeleteAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := getRefreshToken(r)
		if err != nil {
			h.log.Warn("refresh token cookie not found", zap.Error(err))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, ErrorResponse{"refresh token required"})
			return
		}

		user, err := jwt.VerifyToken(token)
		if err != nil {
			h.log.Warn("invalid refresh token", zap.Error(err))
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, ErrorResponse{"invalid refresh token"})
			return
		}

		userID, err := strconv.ParseInt(user.ID, 10, 64)
		if err != nil {
			h.log.Error("invalid user ID in token", zap.String("id", user.ID))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, ErrorResponse{"invalid user ID"})
			return
		}

		if err := h.repo.DeleteAccount(r.Context(), userID); err != nil {
			h.log.Error("failed to delete account", zap.Error(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, ErrorResponse{"internal server error"})
			return
		}

		clearAuthCookies(w)
		render.Status(r, http.StatusNoContent)
		render.NoContent(w, r)
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
