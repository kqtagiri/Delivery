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

	order, err := domain.NewOrder(email, address)
	if err != nil {
		slog.Error("Service \"CreateOrder\" get next error", "err", err.Error())
		return nil, err
	}

	if err := s.Repo.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil

}

func (s *OrderService) AddItemsToOrder(ctx context.Context, number, menuID int, email string) error {

	return s.Repo.AddItemsToOrder(ctx, number, menuID, email)

}

func (s *OrderService) DeleteItemsFromOrder(ctx context.Context, number, menuID int, email string) error {

	return s.Repo.DeleteItemsFromOrder(ctx, number, menuID, email)

}

func (s *OrderService) CancelOrder(ctx context.Context, number int, email string) error {

	return s.Repo.CancelOrder(ctx, number, email)

}

func (s *OrderService) GetOrder(ctx context.Context, number int, email string) (*domain.Order, error) {

	return s.Repo.GetOrder(ctx, number, email)

}

func (s *OrderService) GetOrderDetails(ctx context.Context, number int, email string) ([]domain.OrderItem, error) {

	return s.Repo.GetOrderDetails(ctx, number, email)

}

func (s *OrderService) ConfirmOrder(ctx context.Context, number int, email string) error {

	user, err := s.Repo.FillUser(ctx, email)
	if err != nil {
		return err
	}

	order, err := s.Repo.GetOrder(ctx, number, email)
	if err != nil {
		return err
	}

	if user.Balance < order.Cost {
		slog.Error("Don't enough money on balance", "email", email)
		return errors.New("don`t enough money on balance")
	}

	newBalance := user.Balance - order.Cost

	return s.Repo.ConfirmOrder(ctx, number, email, newBalance)

}
