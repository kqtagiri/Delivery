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

type RestaurantHandler struct {
	Service *service.RestaurantService
}

func NewRestaurantHandler(service *service.RestaurantService) *RestaurantHandler {

	return &RestaurantHandler{
		Service: service,
	}

}

type RestDTO struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Address     string  `json:"address" binding:"required"`
	Rating      float64 `json:"rating" binding:"required"`
}

func ConvertRestToDTO(rest *domain.Restaurant) *RestDTO {

	return &RestDTO{
		Title:       rest.Title,
		Description: rest.Description,
		Address:     rest.Address,
		Rating:      rest.Rating,
	}

}

type ItemDTOAdmin struct {
	Title       string  `json:"title" binding:"required"`
	RestTitle   string  `json:"restaurant_title" binding:"required"`
	Composition string  `json:"composition" binding:"required"`
	Time        int     `json:"time" binding:"required"`
	Cost        float64 `json:"cost" binding:"required"`
}

func ConvertItemToDTO(item *domain.Item) *ItemDTOAdmin {

	return &ItemDTOAdmin{
		Title:       item.Title,
		RestTitle:   item.RestTitle,
		Composition: item.Composition,
		Time:        item.Time,
		Cost:        item.Cost,
	}

}

// func ConvertDTOToItem(dto *ItemDTO) *domain.Item {

// 	return &domain.Item{
// 		Title:       dto.Title,
// 		RestTitle:   dto.RestTitle,
// 		Composition: dto.Composition,
// 		Cost:        dto.Cost,
// 	}

// }

func (h *RestaurantHandler) CreateRestaurant(c *gin.Context) {

	slog.Info("Handler started \"CreateRestaurant\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	var req RestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(err.Error())
		c.JSON(400, gin.H{"error": err})
		return
	}

	restaurant, err := h.Service.CreateRestaurant(ctx, req.Title, req.Description, req.Address, req.Rating)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTitle) || errors.Is(err, domain.ErrInvalidDescription) || errors.Is(err, domain.ErrInvalidAddress) || errors.Is(err, domain.ErrInvalidRating) || errors.Is(err, domain.ErrWithInsert) {
			c.JSON(400, gin.H{"error": err})
		} else {
			c.JSON(500, gin.H{"error": err})
		}
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"CreateRestaurant\" took a lot of time")
	}

	dto := ConvertRestToDTO(restaurant)

	slog.Info("Handler ended \"CreateRestaurant\" success")

	c.JSON(201, dto)

}

func (h *RestaurantHandler) RestaurantsList(c *gin.Context) {

	slog.Info("Handler started \"RestaurantsList\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	rests, err := h.Service.RestaurantsList(ctx)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"RestaurantsList\" took a lot of time")
	}

	dtos := []RestDTO{}
	for _, rest := range *rests {

		dtos = append(dtos, *ConvertRestToDTO(&rest))

	}

	slog.Info("Handler ended \"RestaurantsList\" success")
	c.JSON(200, dtos)

}

func (h *RestaurantHandler) RestaurantMenu(c *gin.Context) {

	title := c.Param("title")
	slog.Info("Handler started \"RestaurantMenu\" from %s", title)

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	menu, err := h.Service.RestaurantMenu(ctx, title)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"RestaurantMenu\" from %s took a lot of time", title)
	}

	dtos := []ItemDTOAdmin{}
	for _, item := range *menu {

		dtos = append(dtos, *ConvertItemToDTO(&item))

	}

	slog.Info("Handler ended \"RestaurantMenu\" from %s success", title)
	c.JSON(200, dtos)

}

func (h *RestaurantHandler) AddNewItems(c *gin.Context) {

	slog.Info("Handler started \"AddNewItems\"")

	itemsDTO := []ItemDTOAdmin{}
	if err := c.ShouldBindJSON(&itemsDTO); err != nil {
		slog.Error("Handler \"AddNewItems\" get next error when parsing json:%w", err)
		c.JSON(400, gin.H{"error": err})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	for _, dto := range itemsDTO {

		if err := h.Service.AddNewItems(ctx, dto.Title, dto.RestTitle, dto.Composition, dto.Time, dto.Cost); err != nil {
			if errors.Is(err, domain.ErrInvalidTitle) || errors.Is(err, domain.ErrInvalidRestaurantTitle) || errors.Is(err, domain.ErrInvalidComposition) || errors.Is(err, domain.ErrInvalidCost) || errors.Is(err, domain.ErrInvalidTime) || errors.Is(err, domain.ErrWithInsert) {
				c.JSON(400, gin.H{"error": err})
			} else {
				c.JSON(500, gin.H{"error": err})
			}
			return
		}

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"AddNewItems\" took a lot of time")
	}

	slog.Info("Handler ended \"AddNewItems\" success")
	c.Status(201)

}

func (h *RestaurantHandler) DeleteItems(c *gin.Context) {

	slog.Info("Handler started \"DeleteItems\"")

	itemsDTO := []ItemDTOAdmin{}
	if err := c.BindJSON(&itemsDTO); err != nil {
		slog.Error("Handler \"DeleteItems\" get next error when parsing json:%w", err)
		c.JSON(500, gin.H{"error": err})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	for _, dto := range itemsDTO {

		if err := h.Service.DeleteItems(ctx, dto.Title, dto.RestTitle); err != nil {
			if errors.Is(err, domain.ErrWithDelete) {
				c.JSON(400, gin.H{"error": err})
			} else {
				c.JSON(500, gin.H{"error": err})
			}
			return
		}

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"DeleteItems\" took a lot of time")
	}

	slog.Info("Handler ended \"DeleteItems\" success")
	c.Status(200)

}
