package jwt

import (
	"auth/internal/domain"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

const (
	uidClaim   = "uid"
	emailClaim = "email"
)

func NewToken(user *domain.User, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		uidClaim:   user.ID,
		emailClaim: user.Email,
		"exp":      time.Now().Add(duration).Unix(),
	})
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
func VerifyToken(tokenString string) (*domain.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, fmt.Errorf("malformed token")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, fmt.Errorf("invalid signature")
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, fmt.Errorf("token expired")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, fmt.Errorf("token not active yet")
		default:
			return nil, fmt.Errorf("couldn't handle this token: %w", err)
		}
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	uid, ok := claims[uidClaim].(string)
	if !ok {
		return nil, fmt.Errorf("invalid uid claim")
	}

	email, ok := claims[emailClaim].(string)
	if !ok {
		return nil, fmt.Errorf("invalid email claim")
	}

	user := &domain.User{
		ID:    uid,
		Email: email,
	}

	return user, nil
}

// TODO: доделать хэширование
func HashToken(token string) (string, error) {
	return token, nil
}
