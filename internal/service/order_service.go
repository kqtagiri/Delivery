package service

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/repository"
	"errors"
	"log/slog"
)

type OrderService struct {
	Repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {

	return &OrderService{
		Repo: repo,
	}

}

func (s *OrderService) CreateOrder(ctx context.Context, email, address string) (*domain.Order, error) {

	slog.Info("Service started \"CreateOrder\"")

	order, err := domain.NewOrder(email, address)
	if err != nil {
		slog.Error("Service \"CreateOrder\" get next error:%w", err)
		return nil, err
	}

	if err := s.Repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	slog.Info("Service ended \"CreateOrder\" success")
	return order, nil

}

func (s *OrderService) AddItemsToOrder(ctx context.Context, number int, title, restTitle string) error {

	slog.Info("Service started \"AddItemsToOrder\"")

	if err := s.Repo.AddItemsToOrder(ctx, number, title, restTitle); err != nil {
		return err
	}

	slog.Info("Service ended \"AddItemsToOrder\" success")
	return nil

}

func (s *OrderService) DeleteItemsFromOrder(ctx context.Context, number int, title, restTitle string) error {

	slog.Info("Service started \"DeleteItemsFromOrder\"")

	if err := s.Repo.DeleteItemsFromOrder(ctx, number, title, restTitle); err != nil {
		return err
	}

	slog.Info("Service ended \"DeleteItemsFromOrder\" success")
	return nil

}

func (s *OrderService) CancelOrder(ctx context.Context, number int) error {

	slog.Info("Service started \"CancelOrder\"")

	if err := s.Repo.CancelOrder(ctx, number); err != nil {
		return err
	}

	slog.Info("Service ended \"CancelOrder\" success")
	return nil

}

func (s *OrderService) GetOrder(ctx context.Context, number int) (*domain.Order, error) {

	slog.Info("Service started \"GetOrder\"")

	order, err := s.Repo.GetOrder(ctx, number)
	if err != nil {
		return nil, err
	}

	slog.Info("Service ended \"GetOrder\" success")
	return order, nil

}

func (s *OrderService) GetOrderDetails(ctx context.Context, number int) (*[]domain.Item, error) {

	slog.Info("Service started \"GetOrderDetails\"")

	items, err := s.Repo.GetOrderDetails(ctx, number)
	if err != nil {
		return nil, err
	}

	slog.Info("Service ended \"GetOrderDetails\" success")
	return items, nil

}

func (s *OrderService) ConfirmOrder(ctx context.Context, number int, email string) error {

	slog.Info("Service started \"ConfirmOrder\"")

	user, err := s.Repo.FillUser(ctx, email)
	if err != nil {
		return err
	}

	order, err := s.Repo.GetOrder(ctx, number)
	if err != nil {
		return err
	}

	if user.Balance < order.Cost {
		slog.Error("Cost > Balance, user - %s", email)
		return errors.New("Don`t enough money on balance!")
	}

	newBalance := user.Balance - order.Cost
	if err := s.Repo.ConfirmOrder(ctx, number, email, newBalance); err != nil {
		return err
	}

	slog.Info("Service ended \"ConfirmOrder\" success")
	return nil

}
