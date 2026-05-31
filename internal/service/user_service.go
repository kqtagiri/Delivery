package service

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/repository"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-jwt/jwt"
)

type UserService struct {
	Repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {

	return &UserService{
		Repo: repo,
	}

}

func (s *UserService) RegisterAccount(ctx context.Context, name, email, password, role, address string) (*domain.User, error) {

	slog.Info("Service started \"RegisterAccount\"")

	user, err := domain.NewUser(name, email, password, role, address)
	if err != nil {
		slog.Error("Service \"RegisterAccount\" get next err:%w", err)
		return nil, err
	}

	if err := s.Repo.RegisterAccount(ctx, user); err != nil {
		return nil, err
	}

	slog.Info("Service ended \"RegisterAccount\" success")
	return user, nil

}

func (s *UserService) Authorization(ctx context.Context, email, password string) (*domain.Tokens, error) {

	slog.Info("Service started \"Authorization\"")

	tokens, err := s.Repo.Authorization(ctx, email, password)
	if err != nil {
		return nil, err
	}

	slog.Info("Service ended \"Authorization\" success")
	return tokens, nil

}

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (*domain.Tokens, error) {

	slog.Info("Service started \"Refresh\"")

	claims := &domain.RefreshTokenClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid singing method:%v", t.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET_KEY")), nil
	})

	if err != nil {
		slog.Error("Service \"Refresh\" get next error:", err)
		return nil, err
	}

	if !token.Valid {
		slog.Error("Service \"Refresh\" get next error:", domain.ErrInvalidRefreshToken)
		return nil, domain.ErrInvalidRefreshToken
	}

	tokens, err := s.Repo.Refresh(ctx, claims.UserId)
	if err != nil {
		return nil, err
	}

	slog.Info("Service ended \"Refresh\" success")
	return tokens, nil

}

func (s *UserService) ReplenishBalance(ctx context.Context, balance float64, email string) (*domain.User, error) {

	slog.Info("Service started \"ReplenishBalance\"")

	user, err := s.Repo.UserInfo(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := user.ReplenishBalance(balance); err != nil {
		slog.Error("Service \"ReplenishBalance\" with email %s get next err:%w", email, err)
		return nil, err
	}

	if err := s.Repo.ReplenishBalance(ctx, user.Balance, email); err != nil {
		return nil, err
	}

	slog.Info("Service ended \"ReplenishBalance\" success")
	return user, nil

}

func (s *UserService) UserInfo(ctx context.Context, email string) (*domain.User, error) {

	slog.Info("Service started \"UserInfo\"")

	user, err := s.Repo.UserInfo(ctx, email)
	if err != nil {
		return nil, err
	}

	slog.Info("Service ended \"UserInfo\" success")
	return user, nil

}

func (s *UserService) AllUsersInfo(ctx context.Context) (*[]domain.User, error) {

	slog.Info("Service started \"AllUsersInfo\"")

	users, err := s.Repo.AllUsersInfo(ctx)
	if err != nil {
		return nil, err
	}

	slog.Info("Service ended \"AllUsersInfo\" success")
	return users, nil

}

/*func (s *UserService) UserHistoryInfo(ctx *context.Context) {

}*/
