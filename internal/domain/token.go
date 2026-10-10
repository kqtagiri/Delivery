package domain

import (
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type AccessTokenClaims struct {
	UserId int
	Email  string
	Role   string
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	UserId int
	jwt.StandardClaims
}

func NewToken(accessToken, refreshToken string) (*Tokens, error) {

	if accessToken == "" {
		return nil, ErrInvalidAccessToken
	}

	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	return &Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

func GenerateTokens(id int, email, role string) (*Tokens, error) {

	slog.Info("Started generate token to user", "email", email)

	accessClaims := AccessTokenClaims{
		UserId: id,
		Email:  email,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	accessSecret := os.Getenv("ACCESS_SECRET_KEY")
	if accessSecret == "" {
		slog.Error("Access secret key not found")
		return nil, ErrInvalidAccessToken
	}
	accessTokenStr, err := accessToken.SignedString([]byte(accessSecret))
	if err != nil {
		slog.Error("Get next error when signed access token to user", "email", email)
		return nil, err
	}

	refreshClaims := RefreshTokenClaims{
		UserId: id,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 14).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	refreshSecret := os.Getenv("REFRESH_SECRET_KEY")
	if refreshSecret == "" {
		slog.Error("refresh secret key not found")
		return nil, ErrInvalidRefreshToken
	}
	refreshTokenStr, err := refreshToken.SignedString([]byte(refreshSecret))
	if err != nil {
		slog.Error("Get next error when signed refresh token to user", "email", email)
		return nil, err
	}

	slog.Info("Ended generate token to user", "email", email)

	return NewToken(accessTokenStr, refreshTokenStr)

}
