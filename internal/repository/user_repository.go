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

type UserModel struct {
	Id      int
	Name    string
	Email   string
	Address string
	Balance float64
}

func ConvertModelToUser(model *UserModel) *domain.User {

	return &domain.User{
		Name:    model.Name,
		Email:   model.Email,
		Address: model.Address,
		Balance: model.Balance,
	}

}

type UserRepositoryStruct struct {
	Mtx sync.RWMutex
	DB  *database.DB
}

func NewUserRepositoryStruct(db *database.DB) *UserRepositoryStruct {

	return &UserRepositoryStruct{
		Mtx: sync.RWMutex{},
		DB:  db,
	}

}

type UserRepository interface {
	RegisterAccount(ctx context.Context, u *domain.User) error
	ReplenishBalance(ctx context.Context, balance float64, email string) error
	UserInfo(ctx context.Context, email string) (*domain.User, error)
	AllUsersInfo(ctx context.Context) (*[]domain.User, error)
	//UserHistoryInfo(ctx *context.Context, email string) error
}

func (r *UserRepositoryStruct) CreateUserTable() error {

	slog.Info("Started create user table")

	query := `CREATE TABLE IF NOT EXISTS users(
		id SERIAL PRIMARY KEY,
		name VARCHAR(50) NOT NULL,
		email VARCHAR(50) NOT NULL,
		address VARCHAR(200) NOT NULL,
		balance DECIMAL(9,2),
		UNIQUE(email)
	);`

	_, err := r.DB.Conn.Exec(r.DB.Ctx, query)
	if err != nil {
		return err
	}

	slog.Info("Create user table is success")
	return nil

}

func (r *UserRepositoryStruct) RegisterAccount(ctx context.Context, u *domain.User) error {

	slog.Info("Repository started \"RegisterAccount\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `INSERT INTO users (name, email, address, balance) VALUES ($1,$2,$3,$4);`
	result, err := r.DB.Conn.Exec(r.DB.Ctx, query, u.Name, u.Email, u.Address, u.Balance)
	if err != nil {
		slog.Error("Repository \"RegisterAccount\" get next error:%w", err)
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"RegisterAccount\" get next error:%w", domain.ErrWithInsert)
		return domain.ErrWithInsert
	}

	slog.Info("Repository ended \"RegisterAccount\" success")
	return nil

}

func (r *UserRepositoryStruct) ReplenishBalance(ctx context.Context, balance float64, email string) error {

	slog.Info("Repository started \"ReplenishBalance\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `UPDATE users SET balance = $1 WHERE email = $2;`
	result, err := r.DB.Conn.Exec(r.DB.Ctx, query, balance, email)
	if err != nil {
		slog.Error("Repository \"ReplenishBalance\" with email %s get next error:%w", email, err)
		return err
	}

	affected := result.RowsAffected()
	if affected == 0 {
		return domain.ErrWithUpdate
	}

	slog.Info("Repository ended \"ReplenishBalance\" success")
	return nil

}

func (r *UserRepositoryStruct) UserInfo(ctx context.Context, email string) (*domain.User, error) {

	slog.Info("Repository started \"UserInfo\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := "SELECT * FROM users WHERE email = $1;"
	var model UserModel
	if err := r.DB.Conn.QueryRow(r.DB.Ctx, query, email).Scan(&model.Id, &model.Name, &model.Email, &model.Address, &model.Balance); err != nil {
		slog.Error("Repository \"UserInfo\" with email %s get next error:%w", email, err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	user := ConvertModelToUser(&model)

	slog.Info("Repository ended \"UserInfo\" success")
	return user, nil

}

func (r *UserRepositoryStruct) AllUsersInfo(ctx context.Context) (*[]domain.User, error) {

	slog.Info("Repository started \"AllUsersInfo\"")

	r.Mtx.RLock()
	defer r.Mtx.RUnlock()

	query := `SELECT * FROM users;`
	rows, err := r.DB.Conn.Query(r.DB.Ctx, query)
	if err != nil {
		slog.Error("Repository \"AllUsersInfo\" get next error:%w", err)
		return nil, err
	}
	defer rows.Close()

	var model UserModel
	users := []domain.User{}
	for rows.Next() {

		if err := rows.Scan(&model.Id, &model.Name, &model.Email, &model.Address, &model.Balance); err != nil {
			slog.Error("Repository \"AllUsersInfo\" get next error:%w", err)
			return nil, err
		}

		users = append(users, *ConvertModelToUser(&model))

	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository \"AllUsersInfo\" get next error:%w", err)
		return nil, err
	}

	slog.Info("Repository ended \"AllUsersInfo\" success")
	return &users, nil

}

/*func (r *UserRepositoryStruct) UserHistoryInfo() error {

	slog.Info("Repository start register")

	return nil

}*/
