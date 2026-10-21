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

func (s *RestaurantService) CreateRestaurant(ctx context.Context, title, desc, addr string, rating float64) (*domain.Restaurant, error) {

	rest, err := domain.NewRestaurant(title, desc, addr, rating)
	if err != nil {
		slog.Info("Service \"CreateRestaurant\" get next error", "err", err.Error())
		return nil, err
	}

	if err := s.Repo.CreateRestaurant(ctx, rest); err != nil {
		return nil, err
	}

	return rest, nil

}

func (s *RestaurantService) RestaurantsList(ctx context.Context) (*[]domain.Restaurant, error) {

	return s.Repo.RestaurantsList(ctx)

}

func (s *RestaurantService) RestaurantMenu(ctx context.Context, restID int) ([]domain.Item, error) {

	return s.Repo.RestaurantMenu(ctx, restID)

}

func (s *RestaurantService) AddNewItems(ctx context.Context, restID int, title, composition string, time int, cost float64) error {

	item, err := domain.NewItem(title, composition, time, cost)
	if err != nil {
		slog.Info("Service \"AddNewItems\" get next error when create a new item", "err", err.Error())
		return err
	}

	return s.Repo.AddNewItems(ctx, restID, item)

}

func (s *RestaurantService) DeleteItems(ctx context.Context, restID int, title string) error {

	return s.Repo.DeleteItems(ctx, restID, title)

}
