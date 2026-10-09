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
	Id       int
	Name     string
	Email    string
	Password string
	Role     string
	Address  string
	Balance  float64
}

func ConvertModelToUser(model *UserModel) *domain.User {

	return &domain.User{
		Name:     model.Name,
		Email:    model.Email,
		Role:     model.Role,
		Password: model.Password,
		Address:  model.Address,
		Balance:  model.Balance,
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
	Authorization(ctx context.Context, email, password string) (*domain.Tokens, error)
	Refresh(ctx context.Context, userId int) (*domain.Tokens, error)
	ReplenishBalance(ctx context.Context, balance float64, email string) error
	UserInfo(ctx context.Context, email string) (*domain.User, error)
	AllUsersInfo(ctx context.Context) (*[]domain.User, error)
	//UserHistoryInfo(ctx *context.Context, email string) error
}

func (r *UserRepositoryStruct) RegisterAccount(ctx context.Context, u *domain.User) error {

	slog.Info("Repository started \"RegisterAccount\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `INSERT INTO users (name, email, password, role, address, balance) VALUES ($1,$2,$3,$4,$5,$6);`
	result, err := r.DB.Conn.Exec(ctx, query, u.Name, u.Email, u.Password, u.Role, u.Address, u.Balance)
	if err != nil {
		slog.Error("Repository \"RegisterAccount\" get next error:", err)
		return err
	}

	affected := result.RowsAffected()
	if affected != 1 {
		slog.Error("Repository \"RegisterAccount\" get next error:", domain.ErrWithInsert)
		return domain.ErrWithInsert
	}

	slog.Info("Repository ended \"RegisterAccount\" success")
	return nil

}

func (r *UserRepositoryStruct) Authorization(ctx context.Context, email, password string) (*domain.Tokens, error) {

	slog.Info("Repository started \"Authorization\"")

	query := `SELECT * FROM users WHERE email = $1 AND password = $2;`
	var dto UserModel
	if err := r.DB.Conn.QueryRow(ctx, query, email, password).Scan(&dto.Id, &dto.Name, &dto.Email, &dto.Password, &dto.Role, &dto.Address, &dto.Balance); err != nil {
		slog.Error("Repository \"Authorization\" get next error:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidAuthorization
		}
		return nil, err
	}

	tokens, err := domain.GenerateTokens(dto.Id, dto.Email, dto.Role)
	if err != nil {
		return nil, err
	}

	slog.Info("Repository ended \"Authorization\" success")
	return tokens, nil

}

func (r *UserRepositoryStruct) Refresh(ctx context.Context, userId int) (*domain.Tokens, error) {

	slog.Info("Repository started \"Refresh\"")
	var id int
	var email, role string
	query := `SELECT id, email, role FROM users WHERE id = $1;`
	if err := r.DB.Conn.QueryRow(ctx, query, userId).Scan(&id, &email, &role); err != nil {
		slog.Error("Repository \"Refresh\" get next error:", err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	tokens, err := domain.GenerateTokens(id, email, role)
	if err != nil {
		return nil, err
	}

	slog.Info("Repository ended \"Refresh\" success")
	return tokens, nil

}

func (r *UserRepositoryStruct) ReplenishBalance(ctx context.Context, balance float64, email string) error {

	slog.Info("Repository started \"ReplenishBalance\"")

	r.Mtx.Lock()
	defer r.Mtx.Unlock()

	query := `UPDATE users SET balance = $1 WHERE email = $2;`
	result, err := r.DB.Conn.Exec(ctx, query, balance, email)
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
	if err := r.DB.Conn.QueryRow(ctx, query, email).Scan(&model.Id, &model.Name, &model.Email, &model.Password, &model.Role, &model.Address, &model.Balance); err != nil {
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
	rows, err := r.DB.Conn.Query(ctx, query)
	if err != nil {
		slog.Error("Repository \"AllUsersInfo\" get next error:", err)
		return nil, err
	}
	defer rows.Close()

	var model UserModel
	users := []domain.User{}
	for rows.Next() {

		if err := rows.Scan(&model.Id, &model.Name, &model.Email, &model.Password, &model.Role, &model.Address, &model.Balance); err != nil {
			slog.Error("Repository \"AllUsersInfo\" get next error:", err)
			return nil, err
		}

		users = append(users, *ConvertModelToUser(&model))

	}

	if err := rows.Err(); err != nil {
		slog.Error("Repository \"AllUsersInfo\" get next error:", err)
		return nil, err
	}

	slog.Info("Repository ended \"AllUsersInfo\" success")
	return &users, nil

}

/*func (r *UserRepositoryStruct) UserHistoryInfo() error {

	slog.Info("Repository start register")

	return nil

}*/
