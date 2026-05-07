package repository

import (
	"context"
	"database/sql"
	"delivery/internal/database"
	"delivery/internal/domain"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	AddItemsToOrder(ctx context.Context, number int, title, restTitle string) error
	DeleteItemsFromOrder(ctx context.Context, number int, title, restTitle string) error
	CancelOrder(ctx context.Context, number int) error
	GetOrder(ctx context.Context, number int) (*domain.Order, error)
	GetOrderDetails(ctx context.Context, number int) (*[]domain.Item, error)
	ConfirmOrder(ctx context.Context, number int, email string, newBalance float64) error
	FillUser(ctx context.Context, email string) (*domain.User, error)
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
		number SERIAL,
		email VARCHAR(100) NOT NULL,
		address VARCHAR(300) NOT NULL,
		status VARCHAR(15) NOT NULL,
		time INT NOT NULL,
		cost DECIMAL(9,2) NOT NULL,
		UNIQUE(email)
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

func (r *OrderRepositoryStruct) CreateOrder(ctx context.Context, order *domain.Order) error {

	slog.Info("Repository started \"CreateOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.DB.Conn.Begin(r.DB.Ctx)
	if err != nil {
		slog.Error("Repository \"CreateOrder\" get next error when create a transaction:", err)
		return err
	}

	query := `INSERT INTO orders(email, address, status, time, cost) VALUES($1,$2,$3,$4,$5,$6);`

	result, err := tx.Exec(r.DB.Ctx, query, order.Email, order.Address, order.Status, order.Time, order.Cost)
	if err != nil {
		slog.Error("Repository \"CreateOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"CreateOrder\" have error with insert")
		tx.Rollback(r.DB.Ctx)
		return domain.ErrWithInsert
	}

	query = `SELECT number FROM orders WHERE email = $1;`

	if err := tx.QueryRow(r.DB.Ctx, query, order.Email).Scan(&order.Number); err != nil {
		slog.Error("Repository \"CreateOrder\" get next error:", err)
		tx.Rollback(r.DB.Ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	slog.Info("Repository ended \"CreateOrder\" success")
	return tx.Commit(r.DB.Ctx)

}

func (r *OrderRepositoryStruct) AddItemsToOrder(ctx context.Context, number int, title, restTitle string) error {

	slog.Info("Repository started \"AddItemsToOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.DB.Conn.Begin(r.DB.Ctx)
	if err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error when create a transaction:%w", err)
		return err
	}

	query := `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := tx.QueryRow(r.DB.Ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if status != domain.StatusCreated {
		slog.Error("Repository \"AddItemsToOrder\" try to change order which status not created!")
		tx.Rollback(r.DB.Ctx)
		return fmt.Errorf("Status not created!")
	}

	query = `SELECT * FROM menu WHERE title = $1 AND rest_title = $2;`
	var model ItemModel
	if err := tx.QueryRow(r.DB.Ctx, query, title, restTitle).Scan(&model.Id, &model.Title, &model.RestTitle, &model.Composition, &model.Time, &model.Cost); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrItemNotFound
		}
		return err
	}

	query = `INSERT INTO orders_items(number, title, rest_title, cost, time) VALUES($1,$2,$3,$4,$5);`
	result, err := tx.Exec(r.DB.Ctx, query, number, model.Title, model.RestTitle, model.Cost, model.Time)
	if err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"AddItemsToOrder\" get error with insert")
		tx.Rollback(r.DB.Ctx)
		return domain.ErrWithInsert
	}

	query = `UPDATE orders SET time = time + $1 AND cost = cost + $2 WHERE number = $3;`
	result, err = tx.Exec(r.DB.Ctx, query, model.Time, model.Cost, number)
	if err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	affected = result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"AddItemsToOrder\" get error with update")
		tx.Rollback(r.DB.Ctx)
		return domain.ErrWithUpdate
	}

	slog.Info("Repository ended \"AddItemsToOrder\" success")
	return tx.Commit(r.DB.Ctx)

}

func (r *OrderRepositoryStruct) DeleteItemsFromOrder(ctx context.Context, number int, title, restTitle string) error {

	slog.Info("Repository started \"DeleteItemsFromOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.DB.Conn.Begin(r.DB.Ctx)
	if err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error when create a transaction:%w", err)
		return err
	}

	query := `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := tx.QueryRow(r.DB.Ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if status != domain.StatusCreated {
		slog.Error("Repository \"DeleteItemsFromOrder\" try to change order which status not created!")
		tx.Rollback(r.DB.Ctx)
		return fmt.Errorf("Status not created!")
	}

	query = `SELECT (time, cost) FROM orders_items WHERE title = $1 AND restTitle = $2 AND number = $3;`

	var time int
	var cost float64
	if err := tx.QueryRow(r.DB.Ctx, query, title, restTitle, number).Scan(&time, &cost); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrItemNotFound
		}
		return err
	}

	query = `DELETE FROM orders_items WHERE title = $1 AND rest_title = $2 AND number = $3;`
	result, err := tx.Exec(r.DB.Ctx, query, title, restTitle, number)
	if err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"DeleteItemsFromOrder\" get error with delete")
		tx.Rollback(r.DB.Ctx)
		return domain.ErrWithDelete
	}

	query = `UPDATE orders SET time = time - $1 AND cost = cost - $2 WHERE number = $3;`

	result, err = tx.Exec(r.DB.Ctx, query, time, cost, number)
	if err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	affected = result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"DeleteItemsFromOrder\" get error with update")
		tx.Rollback(r.DB.Ctx)
		return domain.ErrWithUpdate
	}

	slog.Info("Repository ended \"DeleteItemsFromOrder\" success")
	return tx.Commit(r.DB.Ctx)

}

func (r *OrderRepositoryStruct) CancelOrder(ctx context.Context, number int) error {

	slog.Info("Repository started \"CancelOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	status := domain.StatusCanceled
	query := `UPDATE orders SET status = $1 WHERE number = $2;`
	result, err := r.DB.Conn.Exec(r.DB.Ctx, query, status, number)
	if err != nil {
		slog.Error("Repository \"CancelOrder\" get next error:%w", err)
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"CancelOrder\" get error with update")
		return domain.ErrWithUpdate
	}

	slog.Info("Repository ended \"CancelOrder\" success")
	return nil

}

func (r *OrderRepositoryStruct) GetOrder(ctx context.Context, number int) (*domain.Order, error) {

	slog.Info("Repository started \"GetOrder\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT * FROM orders WHERE number = $1;`
	var model OrderModel
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, number).Scan(&model.Id, &model.Number, &model.Email, &model.Address, &model.Status, &model.Time, &model.Cost); err != nil {
		slog.Error("Repository \"GetOrder\" get next error:%w", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	order := ConvertModelToOrder(&model)

	slog.Info("Repository ended \"GetOrder\" success")
	return order, nil

}

func (r *OrderRepositoryStruct) GetOrderDetails(ctx context.Context, number int) (*[]domain.Item, error) {

	slog.Info("Repository started \"GetOrderDetails\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT * FROM orders_items WHERE number = $1;`
	rows, err := r.DB.Conn.Query(r.DB.Ctx, query, number)
	if err != nil {
		slog.Error("Repository \"GetOrderDetails\" get next error:%w", err)
		return nil, err
	}
	defer rows.Close()

	var model ItemModel
	items := []domain.Item{}
	for rows.Next() {

		if err := rows.Scan(&model.Id, &model.Title, &model.RestTitle, &model.Composition, &model.Time, &model.Cost); err != nil {
			slog.Error("Repository \"GetOrderDetails\" get next error:%w", err)
			return nil, err
		}

		items = append(items, *ConvertModelToItem(&model))

	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository \"GetOrderDetails\" get next error:%w", err)
		return nil, err
	}

	slog.Info("Repository ended \"GetOrderDetails\" success")
	return &items, nil

}

func (r *OrderRepositoryStruct) ConfirmOrder(ctx context.Context, number int, email string, newBalance float64) error {

	slog.Info("Repository started \"ConfirmOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.DB.Conn.Begin(r.DB.Ctx)
	if err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error:%w", err)
		return err
	}

	query := `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := tx.QueryRow(r.DB.Ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if status != domain.StatusCreated {
		slog.Error("Repository \"ConfirmOrder\" try to change order which status not created!")
		tx.Rollback(r.DB.Ctx)
		return domain.ErrStatusNotCreated
	}

	status = domain.StatusCompleted
	query = `UPDATE orders SET status = $1 WHERE number = $2;`
	result, err := tx.Exec(r.DB.Ctx, query, status, number)
	if err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"ConfirmOrder\" get error with update")
		return domain.ErrWithUpdate
	}

	query = `UPDATE users SET balance = $1 WHERE email = $2;`
	result, err = tx.Exec(r.DB.Ctx, query, newBalance, email)
	if err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error:%w", err)
		tx.Rollback(r.DB.Ctx)
		return err
	}

	affected = result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"ConfirmOrder\" get error with update")
		return domain.ErrWithUpdate
	}

	slog.Info("Repository ended \"ConfirmOrder\" success")
	return tx.Commit(r.DB.Ctx)

}

func (r *OrderRepositoryStruct) FillUser(ctx context.Context, email string) (*domain.User, error) {

	query := "SELECT * FROM users WHERE email = $1;"

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	var model UserModel
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, email).Scan(&model.Id, &model.Name, &model.Email, &model.Address, &model.Balance); err != nil {
		slog.Error("Repository \"FillUser\" with email %s get next error:%w", email, err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	user := ConvertModelToUser(&model)

	return user, nil

}
