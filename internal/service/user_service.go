package service

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/repository"
	"log/slog"
)

type UserService struct {
	Repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {

	return &UserService{
		Repo: repo,
	}

}

func (s *UserService) RegisterAccount(ctx context.Context, name, email, address string) (*domain.User, error) {

	slog.Info("Service started \"RegisterAccount\"")

	user, err := domain.NewUser(name, email, address)
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

	err, user := s.Repo.UserInfo(ctx, email)
	if err != nil {
		return err, nil
	}

	slog.Info("Service ended \"UserInfo\" success")
	return nil, user

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
