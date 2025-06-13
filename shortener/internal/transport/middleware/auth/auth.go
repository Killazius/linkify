package auth

import (
	"context"
	"github.com/Killazius/linkify-proto/pkg/api"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"linkify/internal/lib/api/response"
	"net/http"
)

type Client interface {
	ValidateToken(ctx context.Context, in *api.TokenRequest, opts ...grpc.CallOption) (*api.TokenResponse, error)
}

type contextKey string

const (
	userIDKey    contextKey = "userID"
	userEmailKey contextKey = "userEmail"
)

func New(auth Client, log *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				log.Debug("Access token cookie not found")
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error("Unauthorized"))
				return
			}

			resp, err := auth.ValidateToken(r.Context(), &api.TokenRequest{
				Token: cookie.Value,
			})
			if err != nil {
				if status.Code(err) == codes.Unauthenticated {
					log.Debug("Invalid token", zap.Error(err))
					w.WriteHeader(http.StatusUnauthorized)
					render.JSON(w, r, response.Error("Unauthorized"))
					return
				}
				log.Error("Failed to validate token", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, response.Error("Internal server error"))
				return
			}

			if !resp.Valid {
				log.Debug("Token validation failed", zap.String("error", resp.Error))
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, response.Error("Unauthorized"))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, resp.UserId)
			ctx = context.WithValue(ctx, userEmailKey, resp.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
