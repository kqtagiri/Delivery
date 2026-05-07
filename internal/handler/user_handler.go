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
	Name    string  `json:"name" binding:"required"`
	Email   string  `json:"email" binding:"email, required"`
	Address string  `json:"address" binding:"required"`
	Balance float64 `json:"balance"`
}

func ConvertUserToDTO(user *domain.User) *UserDTO {

	return &UserDTO{
		Name:    user.Name,
		Email:   user.Email,
		Address: user.Address,
		Balance: user.Balance,
	}

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

func (h *UserHandler) RegisterAccount(c *gin.Context) {

	slog.Info("Handler started \"RegisterAccount\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	var req UserDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Handler \"RegisterAccount\" get next error when parsing json:%w", err)
		c.JSON(400, gin.H{"error": err})
		return
	}

	user, err := h.Service.RegisterAccount(ctx, req.Name, req.Email, req.Address)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidName) || errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrInvalidAddress) || errors.Is(err, domain.ErrWithInsert) {
			c.JSON(400, gin.H{"error": err})
			return
		}
		c.JSON(500, gin.H{"error": err})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"RegisterAccount\" took a lot of time")
	}

	dto := ConvertUserToDTO(user)

	slog.Info("Handler ended \"RegisterAccount\" success")
	c.JSON(201, dto)

}

func (h *UserHandler) ReplenishBalance(c *gin.Context) {

	slog.Info("Handler started \"ReplenishBalance\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	email := c.Param("email")
	var req ReplenishDTO
	if err := c.ShouldBindJSON(req); err != nil {
		slog.Error("Handler \"ReplenishBalance\" get next error when parsing json:%w", err)
		c.JSON(400, gin.H{"error": err})
		return
	}

	user, err := h.Service.ReplenishBalance(ctx, req.Balance, email)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidReplenish) || errors.Is(err, domain.ErrWithUpdate) {
			c.JSON(400, gin.H{"error": err})
		} else if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(404, gin.H{"error": err})
		} else {
			c.JSON(500, gin.H{"error": err})
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

func (h *UserHandler) UserInfo(c *gin.Context) {

	slog.Info("Handler started \"UserInfo\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	email := c.Param("email")
	user, err := h.Service.UserInfo(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(404, gin.H{"error": err})
		} else {
			c.JSON(500, gin.H{"error": err})
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

func (h *UserHandler) AllUsersInfo(c *gin.Context) {

	slog.Info("Handler started \"AllUsersInfo\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	users, err := h.Service.AllUsersInfo(ctx)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
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

/*func UserHistoryInfo(c *gin.Context) {

}*/
