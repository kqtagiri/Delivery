package repository

import (
	"context"
	"delivery/internal/database"
	"delivery/internal/domain"
	"errors"
	"log/slog"
	"sync"
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
}

func NewRestaurantRepositoryStruct(db *database.DB) *RestaurantRepositoryStruct {

	return &RestaurantRepositoryStruct{
		Mtx: sync.RWMutex{},
		DB:  db,
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

func (r *RestaurantRepositoryStruct) CreateRestaurantTable() error {

	slog.Info("Start create restaurants table")

	query := `CREATE TABLE IF NOT EXISTS restaurants(
		id SERIAL PRIMARY KEY, 
		title VARCHAR(100) NOT NULL,
		description VARCHAR(300),
		address VARCHAR(300) NOT NULL,
		rating DECIMAL(3,2) NOT NULL,
		UNIQUE(title, address)
	);`

	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query); err != nil {
		return err
	}

	slog.Info("Create restaurants table success")
	return nil

}

func (r *RestaurantRepositoryStruct) CreateMenuTable() error {

	slog.Info("Started create menu table")

	query := `CREATE TABLE IF NOT EXISTS menu(
		id SERIAL PRIMARY KEY,
		title VARCHAR(100) NOT NULL,
		rest_title VARCHAR(100) NOT NULL,
		composition VARCHAR(300) NOT NULL,
		time INT NOT NULL,
		cost DECIMAL(9,2) NOT NULL
	);`

	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query); err != nil {
		return err
	}

	slog.Info("Create menu table success")
	return nil

}

func (r *RestaurantRepositoryStruct) CreateRestaurant(ctx context.Context, rest *domain.Restaurant) error {

	slog.Info("Repository started \"CreateRestaurant\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `INSERT INTO restaurants (title, description, address, rating) VALUES($1,$2,$3,$4);`
	result, err := r.DB.Conn.Exec(r.DB.Ctx, query, rest.Title, rest.Description, rest.Address, rest.Rating)
	if err != nil {
		slog.Error("Repository \"CreateRestaurant\" get next error:%w", err)
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"CreateRestaurant\" get next error:%w", domain.ErrWithInsert)
		return domain.ErrWithInsert
	}

	slog.Info("Repository ended \"CreateRestaurant\" success")
	return nil

}

func (r *RestaurantRepositoryStruct) RestaurantsList(ctx context.Context) (*[]domain.Restaurant, error) {

	slog.Info("Repository started \"RestaurantsList\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT * FROM restaurants;`
	rows, err := r.DB.Conn.Query(r.DB.Ctx, query)
	if err != nil {
		slog.Error("Repository \"RestaurantsList\" get next error:%w", err)
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
		slog.Error("Repository \"RestaurantsList\" get next error:%w", err)
		return nil, err
	}

	slog.Info("Repository ended \"RestaurantsList\" success")
	return &rests, nil

}

func (r *RestaurantRepositoryStruct) RestaurantMenu(ctx context.Context, title string) (*[]domain.Item, error) {

	slog.Info("Repository started \"RestaurantMenu\" from %s", title)

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT * FROM menu WHERE rest_title = $1;`
	rows, err := r.DB.Conn.Query(r.DB.Ctx, query, title)
	if err != nil {
		slog.Error("Repository \"RestaurantMenu\" from %s get next error:%w", title, err)
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
		slog.Error("Repository \"RestaurantsMenu\" get next error:%w", err)
		return nil, err
	}

	slog.Info("Repository ended \"RestaurantMenu\" from %s success", title)
	return &items, nil

}

func (r *RestaurantRepositoryStruct) AddNewItems(ctx context.Context, item *domain.Item) error {

	slog.Info("Repository started \"AddNewItems\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `INSERT INTO menu(title, rest_title, composition, time, cost) VALUES($1,$2,$3,$4,$5);`
	result, err := r.DB.Conn.Exec(r.DB.Ctx, query, item.Title, item.RestTitle, item.Composition, item.Time, item.Cost)
	if err != nil {
		slog.Error("Repository \"AddNewItems\" get next error:%w", err)
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"AddNewItems\" get next error:%w", domain.ErrWithInsert)
		return domain.ErrWithInsert
	}

	slog.Info("Repository ended \"AddNewItems\" success")
	return nil

}

func (r *RestaurantRepositoryStruct) DeleteItems(ctx context.Context, title, restTitle string) error {

	slog.Info("Repository started \"DeleteItems\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `DELETE FROM menu WHERE title = $1 AND rest_title = $2;`
	result, err := r.DB.Conn.Exec(r.DB.Ctx, query, title, restTitle)
	if err != nil {
		slog.Error("Repository \"DeleteItems\" get next error:%w", err)
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"DeleteItems\" get next error:%w", domain.ErrWithDelete)
		return domain.ErrWithDelete
	}

	slog.Info("Repository ended \"DeleteItems\" success")
	return nil

}
