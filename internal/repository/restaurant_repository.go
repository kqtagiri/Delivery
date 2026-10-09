package repository

import (
	"context"
	"delivery/internal/database"
	"delivery/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrRestaurantsInfo = errors.New("No have any created restaurants")
)

type RestaurantRepository interface {
	CreateRestaurant(ctx context.Context, rest *domain.Restaurant) error
	RestaurantsList(ctx context.Context) (*[]domain.Restaurant, error)
	RestaurantMenu(ctx context.Context, title string) (*[]domain.Item, error)
	AddNewItems(ctx context.Context, item *domain.Item) error
	DeleteItems(ctx context.Context, title, restTitle string) error
}

type RestaurantRepositoryStruct struct {
	Mtx sync.RWMutex
	DB  *database.DB
	RDB *redis.Client
}

func NewRestaurantRepositoryStruct(db *database.DB, rdb *redis.Client) *RestaurantRepositoryStruct {

	return &RestaurantRepositoryStruct{
		Mtx: sync.RWMutex{},
		DB:  db,
		RDB: rdb,
	}

}

type RestModel struct {
	Id          int
	Title       string
	Description string
	Address     string
	Rating      float64
}

func ConvertModelToRest(model *RestModel) *domain.Restaurant {

	return &domain.Restaurant{
		Title:       model.Title,
		Description: model.Description,
		Address:     model.Address,
		Rating:      model.Rating,
	}

}

type ItemModel struct {
	Id          int
	Title       string
	RestTitle   string
	Composition string
	Time        int
	Cost        float64
}

func ConvertModelToItem(model *ItemModel) *domain.Item {

	return &domain.Item{
		Title:       model.Title,
		RestTitle:   model.RestTitle,
		Composition: model.Composition,
		Time:        model.Time,
		Cost:        model.Cost,
	}

}

func (r *RestaurantRepositoryStruct) CreateRestaurant(ctx context.Context, rest *domain.Restaurant) error {

	slog.Info("Repository started \"CreateRestaurant\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `INSERT INTO restaurants (title, description, address, rating) VALUES($1,$2,$3,$4);`
	result, err := r.DB.Conn.Exec(ctx, query, rest.Title, rest.Description, rest.Address, rest.Rating)
	if err != nil {
		slog.Error("Repository \"CreateRestaurant\" get next error:", err.Error())
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"CreateRestaurant\" get next error:", domain.ErrWithInsert)
		return domain.ErrWithInsert
	}

	slog.Info("Repository ended \"CreateRestaurant\" success")
	return nil

}

func (r *RestaurantRepositoryStruct) RestaurantsList(ctx context.Context) (*[]domain.Restaurant, error) {

	slog.Info("Repository started \"RestaurantsList\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	data, err := r.RDB.Get(ctx, "restaurants_list").Result()
	if err != nil && err != redis.Nil {
		slog.Warn("Repository \"RestaurantsList\" get next error when get restaurants_list from redis:", err.Error())
	} else if err == nil {
		rests := []domain.Restaurant{}
		if err := json.Unmarshal([]byte(data), &rests); err != nil {
			slog.Warn("Repository \"RestaurantsList\" get next error when parsing json in struct:", err.Error())
		} else {
			slog.Info("Repository ended \"RestaurantsList\" by redis")
			return &rests, nil
		}
	}

	query := `SELECT * FROM restaurants;`
	rows, err := r.DB.Conn.Query(ctx, query)
	if err != nil {
		slog.Error("Repository \"RestaurantsList\" get next error:", err.Error())
		return nil, err
	}
	defer rows.Close()

	rests := []domain.Restaurant{}
	for rows.Next() {

		var model RestModel
		rows.Scan(&model.Id, &model.Title, &model.Description, &model.Address, &model.Rating)
		rests = append(rests, *ConvertModelToRest(&model))

	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository \"RestaurantsList\" get next error:", err.Error())
		return nil, err
	}

	data2, err := json.Marshal(rests)
	if err != nil {
		slog.Warn("Repository \"RestaurantsList\" get next error when parsing struct in json:", err.Error())
	} else {
		if err := r.RDB.Set(ctx, "restaurants_list", data2, time.Hour).Err(); err != nil {
			slog.Warn("Repository \"RestaurantsList\" get next error when set rests in redis:", err.Error())
		}
	}

	slog.Info("Repository ended \"RestaurantsList\" success")
	return &rests, nil

}

func (r *RestaurantRepositoryStruct) RestaurantMenu(ctx context.Context, title string) (*[]domain.Item, error) {

	slog.Info(fmt.Sprintf("Repository started \"RestaurantMenu\" from %s", title))

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	key := fmt.Sprint("restaurant_menu:" + title)
	data, err := r.RDB.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		slog.Warn("Repository \"RestaurantMenu\" get next error when get menu from redis:", err.Error())
	} else if err == nil {
		menu := []domain.Item{}
		if err := json.Unmarshal([]byte(data), &menu); err != nil {
			slog.Warn("Repository \"RestaurantMenu\" get next error when parsing json in struct:", err.Error())
		} else {
			slog.Info(fmt.Sprintf("Repository ended \"RestaurantMenu\" from %s by redis:", title))
			return &menu, nil
		}
	}

	query := `SELECT * FROM menu WHERE rest_title = $1;`
	rows, err := r.DB.Conn.Query(ctx, query, title)
	if err != nil {
		slog.Error(fmt.Sprintf("Repository \"RestaurantMenu\" from %s get next error:", title, err.Error()))
		return nil, err
	}
	defer rows.Close()

	var model ItemModel
	items := []domain.Item{}
	for rows.Next() {

		rows.Scan(&model.Id, &model.Title, &model.RestTitle, &model.Composition, &model.Time, &model.Cost)
		items = append(items, *ConvertModelToItem(&model))

	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository \"RestaurantMenu\" get next error:", err.Error())
		return nil, err
	}

	data2, err := json.Marshal(items)
	if err != nil {
		slog.Warn("Repository \"RestaurantMenu\" get next error when parsing struct in json:", err.Error())
	} else {
		r.RDB.Set(ctx, key, data2, time.Second*60*15)
	}

	slog.Info(fmt.Sprintf("Repository ended \"RestaurantMenu\" from %s success", title))
	return &items, nil

}

func (r *RestaurantRepositoryStruct) AddNewItems(ctx context.Context, item *domain.Item) error {

	slog.Info("Repository started \"AddNewItems\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `INSERT INTO menu(title, rest_title, composition, time, cost) VALUES($1,$2,$3,$4,$5);`
	result, err := r.DB.Conn.Exec(ctx, query, item.Title, item.RestTitle, item.Composition, item.Time, item.Cost)
	if err != nil {
		slog.Error("Repository \"AddNewItems\" get next error:", err.Error())
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"AddNewItems\" get next error:", domain.ErrWithInsert)
		return domain.ErrWithInsert
	}

	key := fmt.Sprint("restaurant_menu:" + item.RestTitle)
	if err := r.RDB.Del(ctx, key).Err(); err != nil {
		slog.Warn("Get next error when delete data from redis:", err.Error())
	}

	slog.Info("Repository ended \"AddNewItems\" success")
	return nil

}

func (r *RestaurantRepositoryStruct) DeleteItems(ctx context.Context, title, restTitle string) error {

	slog.Info("Repository started \"DeleteItems\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `DELETE FROM menu WHERE title = $1 AND rest_title = $2;`
	result, err := r.DB.Conn.Exec(ctx, query, title, restTitle)
	if err != nil {
		slog.Error("Repository \"DeleteItems\" get next error:", err)
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"DeleteItems\" get next error:", domain.ErrWithDelete)
		return domain.ErrWithDelete
	}

	key := fmt.Sprint("restaurant_menu:" + restTitle)
	if err := r.RDB.Del(ctx, key).Err(); err != nil {
		slog.Warn("Get next error when delete data from redis:", err.Error())
	}

	slog.Info("Repository ended \"DeleteItems\" success")
	return nil

}
