package repository

import (
	"context"
	"database/sql"
	"delivery/internal/database"
	"delivery/internal/domain"
	"errors"
	"log/slog"
	"sync"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	AddItemsToOrder(ctx context.Context, number int, email, title, restTitle string) error
	DeleteItemsFromOrder(ctx context.Context, number int, email, title, restTitle string) error
	CancelOrder(ctx context.Context, number int, email string) error
	GetOrder(ctx context.Context, number int, email string) (*domain.Order, error)
	GetOrderDetails(ctx context.Context, number int, email string) (*[]domain.Item, error)
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

func (r *OrderRepositoryStruct) CreateOrder(ctx context.Context, order *domain.Order) error {

	slog.Info("Repository started \"CreateOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.DB.Conn.Begin(ctx)
	if err != nil {
		slog.Error("Repository \"CreateOrder\" get next error when create a transaction", "err", err)
		return err
	}

	defer func() {

		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Repository \"CreateOrder\" rollback failed", "err", err)
		}

	}()

	query := `INSERT INTO orders(email, address, status, time, cost) VALUES($1,$2,$3,$4,$5);`

	result, err := tx.Exec(ctx, query, order.Email, order.Address, order.Status, order.Time, order.Cost)
	if err != nil {
		slog.Error("Repository \"CreateOrder\" get next error", "err", err)
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"CreateOrder\" have error with insert")
		return domain.ErrWithInsert
	}

	query = `SELECT number FROM orders WHERE email = $1;`

	if err := tx.QueryRow(ctx, query, order.Email).Scan(&order.Number); err != nil {
		slog.Error("Repository \"CreateOrder\" get next error", "err", err)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	slog.Info("Repository ended \"CreateOrder\" success")
	return tx.Commit(ctx)

}

func (r *OrderRepositoryStruct) AddItemsToOrder(ctx context.Context, number int, email, title, restTitle string) error {

	slog.Info("Repository started \"AddItemsToOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.DB.Conn.Begin(ctx)
	if err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error when create a transaction", "err", err.Error())
		return err
	}

	defer func() {

		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Repository \"AddItemsToOrder\" rollback failed", "err", err)
		}

	}()

	query := `SELECT email FROM orders WHERE number = $1;`
	var emailDB string
	if err := tx.QueryRow(ctx, query, number).Scan(&email); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if email != emailDB {
		slog.Error("Repository \"AddItemsToOrder\" get next error", "err", domain.ErrInvalidNumber)
		return domain.ErrInvalidNumber
	}

	query = `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := tx.QueryRow(ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if status != domain.StatusCreated {
		slog.Error("Repository \"AddItemsToOrder\" try to change order which status not created!")
		return errors.New("status not created")
	}

	query = `SELECT * FROM menu WHERE title = $1 AND rest_title = $2;`
	var model ItemModel
	if err := tx.QueryRow(ctx, query, title, restTitle).Scan(&model.Id, &model.Title, &model.RestTitle, &model.Composition, &model.Time, &model.Cost); err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrItemNotFound
		}
		return err
	}

	query = `INSERT INTO orders_items(number, title, rest_title, composition, cost, time) VALUES($1,$2,$3,$4,$5,$6);`
	result, err := tx.Exec(ctx, query, number, model.Title, model.RestTitle, model.Composition, model.Cost, model.Time)
	if err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error", "err", err.Error())
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"AddItemsToOrder\" get error with insert")
		return domain.ErrWithInsert
	}

	query = `UPDATE orders SET time = time + $1, cost = cost + $2 WHERE number = $3;`
	result, err = tx.Exec(ctx, query, model.Time, model.Cost, number)
	if err != nil {
		slog.Error("Repository \"AddItemsToOrder\" get next error", "err", err.Error())
		return err
	}

	affected = result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"AddItemsToOrder\" get error with update")
		return domain.ErrWithUpdate
	}

	slog.Info("Repository ended \"AddItemsToOrder\" success")
	return tx.Commit(ctx)

}

func (r *OrderRepositoryStruct) DeleteItemsFromOrder(ctx context.Context, number int, email, title, restTitle string) error {

	slog.Info("Repository started \"DeleteItemsFromOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.DB.Conn.Begin(ctx)
	if err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error when create a transaction", "err", err.Error())
		return err
	}

	defer func() {

		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Repository \"DeleteItemsFromOrder\" rollback failed", "err", err)
		}

	}()

	query := `SELECT email FROM orders WHERE number = $1;`
	var emailDB string
	if err := tx.QueryRow(ctx, query, number).Scan(&email); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if email != emailDB {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error", "err", domain.ErrInvalidNumber)
		return domain.ErrInvalidNumber
	}

	query = `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := tx.QueryRow(ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if status != domain.StatusCreated {
		slog.Error("Repository \"DeleteItemsFromOrder\" try to change order which status not created!")
		return errors.New("status not created")
	}

	query = `SELECT time, cost FROM orders_items WHERE title = $1 AND rest_title = $2 AND number = $3;`

	var time int
	var cost float64
	if err := tx.QueryRow(ctx, query, title, restTitle, number).Scan(&time, &cost); err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrItemNotFound
		}
		return err
	}

	query = `DELETE FROM orders_items WHERE title = $1 AND rest_title = $2 AND number = $3;`
	result, err := tx.Exec(ctx, query, title, restTitle, number)
	if err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error", "err", err.Error())
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"DeleteItemsFromOrder\" get error with delete")
		return domain.ErrWithDelete
	}

	query = `UPDATE orders SET time = time - $1, cost = cost - $2 WHERE number = $3;`

	result, err = tx.Exec(ctx, query, time, cost, number)
	if err != nil {
		slog.Error("Repository \"DeleteItemsFromOrder\" get next error", "err", err.Error())
		return err
	}

	affected = result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"DeleteItemsFromOrder\" get error with update")
		return domain.ErrWithUpdate
	}

	slog.Info("Repository ended \"DeleteItemsFromOrder\" success")
	return tx.Commit(ctx)

}

func (r *OrderRepositoryStruct) CancelOrder(ctx context.Context, number int, email string) error {

	slog.Info("Repository started \"CancelOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `SELECT email FROM orders WHERE number = $1;`
	var emailDB string
	if err := r.DB.Conn.QueryRow(ctx, query, number).Scan(&email); err != nil {
		slog.Error("Repository \"CancelOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if email != emailDB {
		slog.Error("Repository \"CancelOrder\" get next error", "err", domain.ErrInvalidNumber)
		return domain.ErrInvalidNumber
	}

	status := domain.StatusCanceled
	query = `UPDATE orders SET status = $1 WHERE number = $2;`
	result, err := r.DB.Conn.Exec(ctx, query, status, number)
	if err != nil {
		slog.Error("Repository \"CancelOrder\" get next error", "err", err.Error())
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

func (r *OrderRepositoryStruct) GetOrder(ctx context.Context, number int, email string) (*domain.Order, error) {

	slog.Info("Repository started \"GetOrder\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT email FROM orders WHERE number = $1;`
	var emailDB string
	if err := r.DB.Conn.QueryRow(ctx, query, number).Scan(&email); err != nil {
		slog.Error("Repository \"GetOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	if email != emailDB {
		slog.Error("Repository \"GetOrder\" get next error", "err", domain.ErrInvalidNumber)
		return nil, domain.ErrInvalidNumber
	}

	query = `SELECT * FROM orders WHERE number = $1;`
	var model OrderModel
	if err := r.DB.Conn.QueryRow(ctx, query, number).Scan(&model.Id, &model.Number, &model.Email, &model.Address, &model.Status, &model.Time, &model.Cost); err != nil {
		slog.Error("Repository \"GetOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	order := ConvertModelToOrder(&model)

	slog.Info("Repository ended \"GetOrder\" success")
	return order, nil

}

func (r *OrderRepositoryStruct) GetOrderDetails(ctx context.Context, number int, email string) (*[]domain.Item, error) {

	slog.Info("Repository started \"GetOrderDetails\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT email FROM orders WHERE number = $1;`
	var emailDB string
	if err := r.DB.Conn.QueryRow(ctx, query, number).Scan(&email); err != nil {
		slog.Error("Repository \"GetOrderDetails\" get next error", "err", err.Error())

		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, err
	}

	if email != emailDB {
		slog.Error("Repository \"GetOrderDetails\" get next error", "err", domain.ErrInvalidNumber)
		return nil, domain.ErrInvalidNumber
	}

	query = `SELECT * FROM orders_items WHERE number = $1;`
	rows, err := r.DB.Conn.Query(ctx, query, number)
	if err != nil {
		slog.Error("Repository \"GetOrderDetails\" get next error", "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	var model ItemModel
	items := []domain.Item{}
	for rows.Next() {

		var number int
		if err := rows.Scan(&model.Id, &number, &model.Title, &model.RestTitle, &model.Composition, &model.Time, &model.Cost); err != nil {
			slog.Error("Repository \"GetOrderDetails\" get next error", "err", err.Error())
			return nil, err
		}

		items = append(items, *ConvertModelToItem(&model))

	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository \"GetOrderDetails\" get next error", "err", err.Error())
		return nil, err
	}

	slog.Info("Repository ended \"GetOrderDetails\" success")
	return &items, nil

}

func (r *OrderRepositoryStruct) ConfirmOrder(ctx context.Context, number int, email string, newBalance float64) error {

	slog.Info("Repository started \"ConfirmOrder\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	tx, err := r.DB.Conn.Begin(ctx)
	if err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error", "err", err.Error())
		return err
	}

	defer func() {

		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("Repository \"ConfirmOrder\" rollback failed", "err", err)
		}

	}()

	query := `SELECT email FROM orders WHERE number = $1;`
	var emailDB string
	if err := tx.QueryRow(ctx, query, number).Scan(&email); err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if email != emailDB {
		slog.Error("Repository \"ConfirmOrder\" get next error", "err", domain.ErrInvalidNumber)
		return domain.ErrInvalidNumber
	}

	query = `SELECT status FROM orders WHERE number = $1;`
	status := ""
	if err := tx.QueryRow(ctx, query, number).Scan(&status); err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error", "err", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	if status != domain.StatusCreated {
		slog.Error("Repository \"ConfirmOrder\" try to change order which status not created!")
		return domain.ErrStatusNotCreated
	}

	status = domain.StatusCompleted
	query = `UPDATE orders SET status = $1 WHERE number = $2;`
	result, err := tx.Exec(ctx, query, status, number)
	if err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error", "err", err.Error())
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"ConfirmOrder\" get error with update")
		return domain.ErrWithUpdate
	}

	query = `UPDATE users SET balance = $1 WHERE email = $2;`
	result, err = tx.Exec(ctx, query, newBalance, email)
	if err != nil {
		slog.Error("Repository \"ConfirmOrder\" get next error", "err", err.Error())
		return err
	}

	affected = result.RowsAffected()
	if affected == 0 {
		slog.Error("Repository \"ConfirmOrder\" get error with update")
		return domain.ErrWithUpdate
	}

	slog.Info("Repository ended \"ConfirmOrder\" success")
	return tx.Commit(ctx)

}

func (r *OrderRepositoryStruct) FillUser(ctx context.Context, email string) (*domain.User, error) {

	query := "SELECT * FROM users WHERE email = $1;"

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	var model UserModel
	if err := r.DB.Conn.QueryRow(ctx, query, email).Scan(&model.Id, &model.Name, &model.Email, &model.Address, &model.Balance); err != nil {
		slog.Error("Repository \"FillUser\" with email %s get next error:", email, err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	user := ConvertModelToUser(&model)

	return user, nil

}
