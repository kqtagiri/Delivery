package service

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/repository"
	"fmt"
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

func (s *OrderService) CreateOrder(ctx *context.Context, email, address string) (error, *domain.Order) {

	slog.Info("Service started \"CreateOrder\"")

	err, order := domain.NewOrder(email, address)
	if err != nil {
		slog.Error("Service \"CreateOrder\" get next error:%w", err)
		return err, nil
	}

	if err := s.Repo.CreateOrder(ctx, order); err != nil {
		return err, nil
	}

	slog.Info("Service ended \"CreateOrder\" success")
	return nil, order

}

func (s *OrderService) AddItemsToOrder(ctx *context.Context, number int, title, restTitle string) error {

	slog.Info("Service started \"AddItemsToOrder\"")

	if err := s.Repo.AddItemsToOrder(ctx, number, title, restTitle); err != nil {
		return err
	}

	slog.Info("Service ended \"AddItemsToOrder\" success")
	return nil

}

func (s *OrderService) DeleteItemsFromOrder(ctx *context.Context, number int, title, restTitle string) error {

	slog.Info("Service started \"DeleteItemsFromOrder\"")

	if err := s.Repo.DeleteItemsFromOrder(ctx, number, title, restTitle); err != nil {
		return err
	}

	slog.Info("Service ended \"DeleteItemsFromOrder\" success")
	return nil

}

func (s *OrderService) DeleteOrder(ctx *context.Context, number int) error {

	slog.Info("Service started \"DeleteOrder\"")

	if err := s.Repo.DeleteOrder(ctx, number); err != nil {
		return err
	}

	slog.Info("Service ended \"DeleteOrder\" success")
	return nil

}

func (s *OrderService) OrderInfo(ctx *context.Context, number int) (error, *domain.Order) {

	slog.Info("Service started \"OrderInfo\"")

	err, order := s.Repo.OrderInfo(ctx, number)
	if err != nil {
		return err, nil
	}

	slog.Info("Service ended \"OrderInfo\" success")
	return nil, order

}

func (s *OrderService) OrderDetailInfo(ctx *context.Context, number int) (error, *[]domain.Item) {

	slog.Info("Service started \"OrderDetailInfo\"")

	err, items := s.Repo.OrderDetailInfo(ctx, number)
	if err != nil {
		return err, nil
	}

	slog.Info("Service ended \"OrderDetailInfo\" success")
	return nil, items

}

func (s *OrderService) ConfirmOrder(ctx *context.Context, number int, email string) error {

	slog.Info("Service started \"ConfirmOrder\"")

	err, user := s.Repo.FillUser(ctx, email)
	if err != nil {
		return err
	}

	err, order := s.Repo.OrderInfo(ctx, number)
	if err != nil {
		return err
	}

	if user.Balance < order.Cost {
		slog.Error("Cost > Balance, user - %s", email)
		return fmt.Errorf("Don`t enough money on balance!")
	}

	newBalance := user.Balance - order.Cost
	if err := s.Repo.ConfirmOrder(ctx, number, email, newBalance); err != nil {
		return err
	}

	slog.Info("Service ended \"ConfirmOrder\" success")
	return nil

}
