package repository

import (
	"context"
	"delivery/internal/database"
	"delivery/internal/domain"
	"fmt"
	"log/slog"
	"sync"
)

type OrderRepository interface {
	CreateOrder(ctx *context.Context, order *domain.Order) error
	AddItemsToOrder(ctx *context.Context, number int, title, restTitle string) error
	DeleteItemsFromOrder(ctx *context.Context, number int, title, restTitle string) error
	DeleteOrder(ctx *context.Context, number int) error
	OrderInfo(ctx *context.Context, number int) (error, *domain.Order)
	OrderDetailInfo(ctx *context.Context, number int) (error, *[]domain.Item)
	ConfirmOrder(ctx *context.Context, number int, email string, newBalance float64) error
	FillUser(ctx *context.Context, email string) (error, *domain.User)
}

type OrderRepositoryStruct struct {
	Mtx sync.RWMutex
	DB  *database.DB
}

func NewOrderRepositoryStruct(db *database.DB) *OrderRepositoryStruct {

	return &OrderRepositoryStruct{
		Mtx: sync.RWMutex{},
		DB:  db,
	}

}

type OrderModel struct {
	Id      int
	Number  int
	Email   string
	Address string
	Status  string
	Time    int
	Cost    float64
}

func ConvertModelToOrder(model *OrderModel) *domain.Order {

	return &domain.Order{
		Number:  model.Number,
		Email:   model.Email,
		Address: model.Address,
		Status:  model.Status,
		Time:    model.Time,
		Cost:    model.Cost,
	}

}

type ItemInOrder struct {
	Id        int
	Number    int
	Title     string
	RestTitle string
	Time      int
	Cost      float64
}

func (r *OrderRepositoryStruct) CreateOrderTable() error {

	slog.Info("Start create orders table")

	query := `CREATE TABLE IF NOT EXISTS orders(
		id SERIAL PRIMARY KEY,
		number INT NOT NULL,
		email VARCHAR(100) NOT NULL,
		address VARCHAR(300) NOT NULL,
		status VARCHAR(15) NOT NULL,
		time INT NOT NULL,
		cost DECIMAL(9,2) NOT NULL,
		UNIQUE(number)  
	);`

	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query); err != nil {
		return err
	}

	slog.Info("Create orders table success")
	return nil

}

func (r *OrderRepositoryStruct) CreateItemsInOrdersTable() error {

	slog.Info("Started create table to order items")

	query := `CREATE TABLE IF NOT EXISTS orders_items(
		id SERIAL PRIMARY KEY,
		number INT NOT NULL,
		title VARCHAR(100) NOT NULL,
		rest_title VARCHAR(100) NOT NULL,
		time INT NOT NULL,
		cost DECIMAL(9,2) NOT NULL
	);`

	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query); err != nil {
		return err
	}

	slog.Info("Create table to order items success")
	return nil

}

func (r *OrderRepositoryStruct) CreateOrder(ctx *context.Context, order *domain.Order) error {

	slog.Info("Repository started \"CreateOrder\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `INSERT INTO orders(number, email, address, status, time, cost) VALUES($1,$2,$3,$4,$5,$6);`

	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, order.Number, order.Email, order.Address, order.Status, order.Time, order.Cost); err != nil {
		slog.Error("Repository \"CreateOrder\" get next error:%w", err)
		return err
	}

	slog.Info("Repository ended \"CreateOrder\" success")
	return nil

}

func (r *OrderRepositoryStruct) AddItemsToOrder(ctx *context.Context, number int, title, restTitle string) error {

	slog.Info("Repository started \"AddItemsToOrder\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	tx, err := r.DB.Conn.Begin(r.DB.Ctx)
	if err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error when create a transaction:%w", err)
		return err
	}

	query := `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	if status != "Created" {
		slog.Error("Repository \"AddItemsToOrder\" try to change order which status not created!")
		tx.Rollback(r.DB.Ctx)
		return fmt.Errorf("Status not created!")
	}

	query = `SELECT * FROM menu WHERE title = $1 AND rest_title = $2;`
	var model ItemModel
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, title, restTitle).Scan(&model.Id, &model.Title, &model.RestTitle, &model.Composition, &model.Time, &model.Cost); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	query = `INSERT INTO orders_items(number, title, rest_title, cost, time) VALUES($1,$2,$3,$4,$5);`
	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, number, model.Title, model.RestTitle, model.Cost, model.Time); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	query = `UPDATE orders SET time = time + $1 AND cost = cost + $2 WHERE number = $3;`
	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, model.Time, model.Cost, number); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	slog.Info("Repository ended \"AddItemsToOrder\" success")
	return tx.Commit(r.DB.Ctx)

}

func (r *OrderRepositoryStruct) DeleteItemsFromOrder(ctx *context.Context, number int, title, restTitle string) error {

	slog.Info("Repository started \"DeleteItemsFromOrder\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	tx, err := r.DB.Conn.Begin(r.DB.Ctx)
	if err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error when create a transaction:%w", err)
		return err
	}

	query := `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	if status != "Created" {
		slog.Error("Repository \"DeleteItemsFromOrder\" try to change order which status not created!")
		tx.Rollback(r.DB.Ctx)
		return fmt.Errorf("Status not created!")
	}

	query = `SELECT (time, cost) FROM orders_items WHERE title = $1 AND restTitle = $2 AND number = $3;`

	var time int
	var cost float64
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, title, restTitle, number).Scan(&time, &cost); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	query = `DELETE FROM orders_items WHERE title = $1 AND rest_title = $2 AND number = $3;`
	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, title, restTitle, number); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	query = `UPDATE orders SET time = time - $1 AND cost = cost - $2 WHERE number = $3;`

	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, time, cost, number); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	slog.Info("Repository ended \"DeleteItemsFromOrder\" success")
	return tx.Commit(r.DB.Ctx)

}

func (r *OrderRepositoryStruct) DeleteOrder(ctx *context.Context, number int) error {

	slog.Info("Repository started \"DeleteOrder\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	//ПОМЕЧАЕТ ОТМЕНЕННЫМИ!
	status := "Canceled"
	query := `UPDATE orders SET status = $1 WHERE number = $2;`
	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, status, number); err != nil {
		slog.Error("Repository \"DeleteOrder\" get next error:%w", err)
		return err
	}

	//УДАЛЯЕТ, А НЕ ПОМЕЧАЕТ ОТМЕНЕННЫМИ!
	/*tx, err := r.DB.Conn.Begin(r.DB.Ctx)
	if err != nil {
		slog.Error("Repository \"DeleteOrder\" get next error when create a transaction:%w", err)
		return fmt.Errorf("get next error:%w", err)
	}

	query := `DELETE FROM orders WHERE number = $1;`

	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, number); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return fmt.Errorf("get next error:%w", err)
	}

	query = `DELETE FROM orders_items WHERE number = $1;`

	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, number); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return fmt.Errorf("get next error:%w", err)
	}

	slog.Info("Repository end \"DeleteOrder\" success")
	return tx.Commit(r.DB.Ctx)*/

	slog.Info("Repository ended \"DeleteOrder\" success")
	return nil

}

func (r *OrderRepositoryStruct) OrderInfo(ctx *context.Context, number int) (error, *domain.Order) {

	slog.Info("Repository started \"OrderInfo\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT * FROM orders WHERE number = $1;`
	var model OrderModel
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, number).Scan(&model.Id, &model.Number, &model.Email, &model.Address, &model.Status, &model.Time, &model.Cost); err != nil {
		slog.Error("Repository \"OrderInfo\" get next error:%w", err)
		return err, nil
	}

	order := ConvertModelToOrder(&model)

	slog.Info("Repository ended \"OrderInfo\" success")
	return nil, order

}

func (r *OrderRepositoryStruct) OrderDetailInfo(ctx *context.Context, number int) (error, *[]domain.Item) {

	slog.Info("Repository started \"OrderDetailInfo\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT * FROM orders_items WHERE number = $1;`
	rows, err := r.DB.Conn.Query(r.DB.Ctx, query, number)
	if err != nil {
		slog.Error("Repository \"OrderDetailInfo\" get next error:%w", err)
		return err, nil
	}

	var model ItemModel
	items := []domain.Item{}
	for rows.Next() {

		if err := rows.Scan(&model.Id, &model.Title, &model.RestTitle, &model.Composition, &model.Time, &model.Cost); err != nil {
			slog.Error("Repository \"OrderDetailInfo\" get next error:%w", err)
			return err, nil
		}

		items = append(items, *ConvertModelToItem(&model))

	}

	slog.Info("Repository ended \"OrderDetailInfo\" success")
	return nil, &items

}

func (r *OrderRepositoryStruct) ConfirmOrder(ctx *context.Context, number int, email string, newBalance float64) error {

	slog.Info("Repository started \"ConfirmOrder\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	tx, err := r.DB.Conn.Begin(r.DB.Ctx)
	if err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error:%w", err)
		return err
	}

	query := `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	if status != "Created" {
		slog.Error("Repository \"ConfirmOrder\" try to change order which status not created!")
		tx.Rollback(r.DB.Ctx)
		return err
	}

	status = "Confirmed"
	query = `UPDATE orders SET status = $1 WHERE number = $2;`
	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, status, number); err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	query = `UPDATE users SET balance = $1 WHERE email = $2;`
	if _, err := r.DB.Conn.Exec(r.DB.Ctx, query, newBalance, email); err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	slog.Info("Repository ended \"ConfirmOrder\" success")
	return tx.Commit(r.DB.Ctx)

}

func (r *OrderRepositoryStruct) FillUser(ctx *context.Context, email string) (error, *domain.User) {

	query := "SELECT * FROM users WHERE email = $1;"

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	var model UserModel
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, email).Scan(&model.Id, &model.Name, &model.Email, &model.Address, &model.Balance); err != nil {
		slog.Error("Repository \"FillUser\" with email %s get next error:%w", email, err)
		return err, nil
	}

	user := ConvertModelToUser(&model)

	return nil, user

}
