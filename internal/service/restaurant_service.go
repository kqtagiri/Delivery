package service

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/repository"
	"log/slog"
)

type RestaurantService struct {
	Repo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository) *RestaurantService {

	return &RestaurantService{
		Repo: repo,
	}

}

func (s *RestaurantService) CreateRestaurant(ctx *context.Context, title, desc, addr string, rating float64) (error, *domain.Restaurant) {

	slog.Info("Service started \"CreateRestaurant\"")

	err, rest := domain.NewRestaurant(title, desc, addr, rating)
	if err != nil {
		slog.Info("Service \"CreateRestaurant\" get next error:%w", err)
		return err, nil
	}

	if err := s.Repo.CreateRestaurant(ctx, rest); err != nil {
		return err, nil
	}

	slog.Info("Service ended \"CreateRestaurant\" success")
	return nil, rest

}

func (s *RestaurantService) RestaurantsList(ctx *context.Context) (error, *[]domain.Restaurant) {

	slog.Info("Service started \"RestaurantsList\"")

	err, rests := s.Repo.RestaurantsList(ctx)
	if err != nil {
		return err, nil
	}

	slog.Info("Service ended \"RestaurantsList\" success")
	return nil, rests

}

func (s *RestaurantService) RestaurantMenu(ctx *context.Context, title string) (error, *[]domain.Item) {

	slog.Info("Service started \"RestaurantMenu\" from %s", title)

	err, items := s.Repo.RestaurantMenu(ctx, title)
	if err != nil {
		return err, nil
	}

	slog.Info("Service ended \"RestaurantMenu\" from %s success", title)
	return nil, items

}

func (s *RestaurantService) AddNewItems(ctx *context.Context, title, restTitle, composition string, time int, cost float64) error {

	slog.Info("Service started \"AddNewItems\"")

	err, item := domain.NewItem(title, composition, restTitle, time, cost)
	if err != nil {
		slog.Info("Service \"AddNewItems\" get next error when create a new item:%w", err)
		return err
	}

	if err := s.Repo.AddNewItems(ctx, item); err != nil {
		return err
	}

	slog.Info("Service ended \"AddNewItems\" success")
	return nil

}

func (s *RestaurantService) DeleteItems(ctx *context.Context, title, restTitle string) error {

	slog.Info("Service started \"DeleteItems\"")

	if err := s.Repo.DeleteItems(ctx, title, restTitle); err != nil {
		return err
	}

	slog.Info("Service ended \"DeleteItems\" success")
	return nil

}
