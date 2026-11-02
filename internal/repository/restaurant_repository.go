package repository

import (
	"context"
	"delivery/internal/database"
	"delivery/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRestaurantsInfo = errors.New("don't have any created restaurants")
)

type RestaurantRepository interface {
	CreateRestaurant(ctx context.Context, rest *domain.Restaurant) error
	RestaurantsList(ctx context.Context) ([]domain.Restaurant, error)
	RestaurantMenu(ctx context.Context, restID int) ([]domain.Item, error)
	AddNewItems(ctx context.Context, restID int, item *domain.Item) error
	DeleteItems(ctx context.Context, restID int, title string) error
}

type RestaurantRepositoryStruct struct {
	DB  *database.DB
	RDB *redis.Client
}

func NewRestaurantRepositoryStruct(db *database.DB, rdb *redis.Client) *RestaurantRepositoryStruct {

	return &RestaurantRepositoryStruct{
		DB:  db,
		RDB: rdb,
	}

}

type RestModel struct {
	ID          int
	Title       string
	Description string
	Address     string
	Rating      float64
}

func ConvertModelToRest(model *RestModel) *domain.Restaurant {

	return &domain.Restaurant{
		ID:          model.ID,
		Title:       model.Title,
		Description: model.Description,
		Address:     model.Address,
		Rating:      model.Rating,
	}

}

type ItemModel struct {
	ID          int
	RestID      int
	Title       string
	Composition string
	Time        int
	Cost        float64
}

func ConvertModelToItem(model *ItemModel) *domain.Item {

	return &domain.Item{
		ID:          model.ID,
		RestID:      model.RestID,
		Title:       model.Title,
		Composition: model.Composition,
		Time:        model.Time,
		Cost:        model.Cost,
	}

}

func (r *RestaurantRepositoryStruct) CreateRestaurant(ctx context.Context, rest *domain.Restaurant) error {

	query := `INSERT INTO restaurants (title, description, address, rating) VALUES($1,$2,$3,$4);`
	result, err := r.DB.Conn.Exec(ctx, query, rest.Title, rest.Description, rest.Address, rest.Rating)
	if err != nil {
		slog.Error("Repository \"CreateRestaurant\" get next error", "err", err.Error())
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"CreateRestaurant\" get next error", "err", domain.ErrWithInsert)
		return domain.ErrWithInsert
	}

	return nil

}

func (r *RestaurantRepositoryStruct) RestaurantsList(ctx context.Context) ([]domain.Restaurant, error) {

	data, err := r.RDB.Get(ctx, "restaurants_list").Result()
	if err != nil && err != redis.Nil {
		slog.Warn("Repository \"RestaurantsList\" get next error when get restaurants_list from redis", "err", err.Error())
	} else if err == nil {
		rests := []domain.Restaurant{}
		if err := json.Unmarshal([]byte(data), &rests); err != nil {
			slog.Warn("Repository \"RestaurantsList\" get next error when parsing json in struct", "err", err.Error())
		} else {
			slog.Info("Repository ended \"RestaurantsList\" by redis")
			return rests, nil
		}
	}

	query := `SELECT * FROM restaurants;`
	rows, err := r.DB.Conn.Query(ctx, query)
	if err != nil {
		slog.Error("Repository \"RestaurantsList\" get next error", "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	rests := []domain.Restaurant{}
	for rows.Next() {

		var model RestModel
		if err := rows.Scan(&model.ID, &model.Title, &model.Description, &model.Address, &model.Rating); err != nil {
			slog.Error("Scan failed", "err", err)
			return nil, err
		}

		rests = append(rests, *ConvertModelToRest(&model))

	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository \"RestaurantsList\" get next error", "err", err.Error())
		return nil, err
	}

	data2, err := json.Marshal(rests)
	if err != nil {
		slog.Warn("Repository \"RestaurantsList\" get next error when parsing struct in json", "err", err.Error())
	} else {
		if err := r.RDB.Set(ctx, "restaurants_list", data2, time.Hour).Err(); err != nil {
			slog.Warn("Repository \"RestaurantsList\" get next error when set rests in redis", "err", err.Error())
		}
	}

	return rests, nil

}

func (r *RestaurantRepositoryStruct) RestaurantMenu(ctx context.Context, restID int) ([]domain.Item, error) {

	key := fmt.Sprint("restaurant_menu:" + strconv.Itoa(restID))
	data, err := r.RDB.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		slog.Warn("Repository \"RestaurantMenu\" get next error when get menu from redis", "err", err.Error())
	} else if err == nil {
		menu := []domain.Item{}
		if err := json.Unmarshal([]byte(data), &menu); err != nil {
			slog.Warn("Repository \"RestaurantMenu\" get next error when parsing json in struct", "err", err.Error())
		} else {
			slog.Info(fmt.Sprintf("Repository ended \"RestaurantMenu\" from %d by redis:", restID))
			return menu, nil
		}
	}

	query := `SELECT * FROM menu WHERE rest_id = $1;`
	rows, err := r.DB.Conn.Query(ctx, query, restID)
	if err != nil {
		slog.Error("Repository \"RestaurantMenu\" get next error", "id", restID, "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	var model ItemModel
	items := []domain.Item{}
	for rows.Next() {

		if err := rows.Scan(&model.ID, &model.RestID, &model.Title, &model.Composition, &model.Time, &model.Cost); err != nil {
			slog.Error("Scan failed", "err", err)
			return nil, err
		}

		items = append(items, *ConvertModelToItem(&model))

	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository \"RestaurantMenu\" get next error", "err", err.Error())
		return nil, err
	}

	data2, err := json.Marshal(items)
	if err != nil {
		slog.Warn("Repository \"RestaurantMenu\" get next error when parsing struct in json", "err", err.Error())
	} else {
		r.RDB.Set(ctx, key, data2, time.Second*60*15)
	}

	return items, nil

}

func (r *RestaurantRepositoryStruct) AddNewItems(ctx context.Context, restID int, item *domain.Item) error {

	query := `INSERT INTO menu(rest_id, title, composition, time, cost) VALUES($1,$2,$3,$4,$5);`
	result, err := r.DB.Conn.Exec(ctx, query, restID, item.Title, item.Composition, item.Time, item.Cost)
	if err != nil {
		slog.Error("Repository \"AddNewItems\" get next error", "err", err.Error())
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"AddNewItems\" get next error", "err", domain.ErrWithInsert)
		return domain.ErrWithInsert
	}

	key := fmt.Sprint("restaurant_menu:" + strconv.Itoa(restID))
	if err := r.RDB.Del(ctx, key).Err(); err != nil {
		slog.Warn("Get next error when delete data from redis", "err", err.Error())
	}

	return nil

}

func (r *RestaurantRepositoryStruct) DeleteItems(ctx context.Context, restID int, title string) error {

	query := `DELETE FROM menu WHERE rest_id = $1 AND title = $2;`
	result, err := r.DB.Conn.Exec(ctx, query, restID, title)
	if err != nil {
		slog.Error("Repository \"DeleteItems\" get next error", "err", err.Error())
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"DeleteItems\" get next error", "err", domain.ErrWithDelete)
		return domain.ErrWithDelete
	}

	key := fmt.Sprint("restaurant_menu:" + strconv.Itoa(restID))
	if err := r.RDB.Del(ctx, key).Err(); err != nil {
		slog.Warn("Get next error when delete data from redis", "err", err.Error())
	}

	return nil

}
