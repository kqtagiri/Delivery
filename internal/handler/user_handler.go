package handler

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/service"
	"errors"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

type UserDTO struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required"`
	Role     string  `json:"role" binding:"required"`
	Address  string  `json:"address" binding:"required"`
	Balance  float64 `json:"balance"`
}

func ConvertUserToDTO(user *domain.User) *UserDTO {

	return &UserDTO{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
		Role:     user.Role,
		Address:  user.Address,
		Balance:  user.Balance,
	}

}

type UserAuth struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ResponseWithToken struct {
	AccessToken string `json:"access_token"`
}

type ReplenishDTO struct {
	Balance float64 `json:"balance" binding:"required"`
}

type UserHandler struct {
	Service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {

	return &UserHandler{
		Service: service,
	}

}

// @Summary Register account
// @Description Создание нового аккаунта
// @Tags users
// @Router /users/register [post]
// @Accept json
// @Produce json
// @Param body body UserDTO true "Имя, почта, пароль, роль и адрес"
// @Success 201 {object} UserDTO "Регистрация пройдена успешно"
// @Failure 400 {object} domain.ErrorResponse "Неверные входные данные"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
func (h *UserHandler) RegisterAccount(c *gin.Context) {

	slog.Info("Handler started \"RegisterAccount\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	var req UserDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Handler \"RegisterAccount\" get next error when parsing json:%w", err)
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.Service.RegisterAccount(ctx, req.Name, req.Email, req.Password, req.Role, req.Address)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidName) || errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrInvalidPassword) || errors.Is(err, domain.ErrInvalidRole) || errors.Is(err, domain.ErrInvalidAddress) || errors.Is(err, domain.ErrWithInsert) {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"RegisterAccount\" took a lot of time")
	}

	dto := ConvertUserToDTO(user)

	slog.Info("Handler ended \"RegisterAccount\" success")
	c.JSON(201, dto)

}

// @Summary Login
// @Description Авторизация
// @Tags users
// @Router /users/login [post]
// @Accept json
// @Produce json
// @Param body body UserAuth true "Почта и пароль"
// @Success 200 {object} ResponseWithToken "Ваш токен для прохождения авторизации"
// @Failure 400 {object} domain.ErrorResponse "Неверные входные данные"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
func (h *UserHandler) Authorization(c *gin.Context) {

	slog.Info("Handler started \"Authorization\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	var auth UserAuth
	if err := c.ShouldBindJSON(&auth); err != nil {
		slog.Error("Handler \"Authorization\" get next error when parsing json:", err)
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	tokens, err := h.Service.Authorization(ctx, auth.Email, auth.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAuthorization) {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.SetCookie("refresh_token", tokens.RefreshToken, 60*60*24*14, "/users/refresh", "", true, true)

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"Authorization\" took a lot of time")
	}

	slog.Info("Handler ended \"Authorization\" success")
	c.JSON(200, ResponseWithToken{AccessToken: tokens.AccessToken})

}

// @Summary Refresh
// @Description Обновление токенов
// @Tags users
// @Router /users/refresh [get]
// @Produce json
// @Success 200 {object} ResponseWithToken "Ваш новый токен для прохождения авторизации"
// @Failure 400 {object} domain.ErrorResponse "Неверные входные данные"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
func (h *UserHandler) Refresh(c *gin.Context) {

	slog.Info("Handler started \"Refresh\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	start := time.Now()

	refresh, err := c.Cookie("refresh_token")
	if err != nil {
		slog.Error("Handler \"Refresh\" get error when trying to get token from cookie")
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	tokens, err := h.Service.Refresh(ctx, refresh)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAuthorization) {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"Refresh\" took a lot of time")
	}

	slog.Info("Handler ended \"Refresh\" success")
	c.JSON(200, ResponseWithToken{AccessToken: tokens.AccessToken})

}

// @Summary Replenish balance
// @Description Пополнение баланса
// @Tags users
// @Router /users/replenish [patch]
// @Accept json
// @Produce json
// @Param body body ReplenishDTO true "Сумма для пополнения"
// @Success 200 {object} UserDTO "Баланс успешно пополнен"
// @Failure 400 {object} domain.ErrorResponse "Неверные входные данные"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 404 {object} domain.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *UserHandler) ReplenishBalance(c *gin.Context) {

	slog.Info("Handler started \"ReplenishBalance\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	value, exists := c.Get("email")
	if !exists {
		slog.Error("Handler \"ReplenishBalance\" email not found in context")
		c.JSON(401, domain.ErrorResponse{Error: "Email not found"})
		return
	}
	email, ok := value.(string)
	if !ok {
		slog.Error("Handler \"ReplenishBalance\" get error when parsing value in string")
		c.JSON(500, domain.ErrorResponse{Error: "Something went wrong"})
		return
	}

	var req ReplenishDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Handler \"ReplenishBalance\" get next error when parsing json:%w", err)
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.Service.ReplenishBalance(ctx, req.Balance, email)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidReplenish) || errors.Is(err, domain.ErrWithUpdate) {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		} else if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"ReplenishBalance\" took a lot of time")
	}

	dto := ConvertUserToDTO(user)

	slog.Info("Handler ended \"ReplenishBalance\" success")
	c.JSON(200, dto)

}

// @Summary User info
// @Description Информация о пользователе
// @Tags users
// @Router /users [get]
// @Produce json
// @Success 200 {object} UserDTO "OK"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 404 {object} domain.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *UserHandler) UserInfo(c *gin.Context) {

	slog.Info("Handler started \"UserInfo\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	value, exists := c.Get("email")
	if !exists {
		slog.Error("Handler \"ReplenishBalance\" email not found in context")
		c.JSON(401, domain.ErrorResponse{Error: "Email not found"})
		return
	}
	email, ok := value.(string)
	if !ok {
		slog.Error("Handler \"ReplenishBalance\" get error when parsing value in string")
		c.JSON(500, domain.ErrorResponse{Error: "Something went wrong"})
		return
	}

	user, err := h.Service.UserInfo(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"UserInfo\" took a lot of time")
	}

	dto := ConvertUserToDTO(user)

	slog.Info("Handler ended \"UserInfo\" success")
	c.JSON(200, dto)

}

// @Summary All users info
// @Description Информация о всех пользователях
// @Tags users
// @Router /users/all [get]
// @Produce json
// @Success 200 {array} UserDTO "OK"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *UserHandler) AllUsersInfo(c *gin.Context) {

	slog.Info("Handler started \"AllUsersInfo\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	users, err := h.Service.AllUsersInfo(ctx)
	if err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	dtos := []UserDTO{}
	for _, user := range *users {

		dtos = append(dtos, *ConvertUserToDTO(&user))

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"AllUsersInfo\" took a lot of time")
	}

	slog.Info("Handler ended \"AllUsersInfo\" success")
	c.JSON(200, dtos)

}
