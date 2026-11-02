package handler

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/service"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
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
	ID          int     `json:"id"`
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Address     string  `json:"address" binding:"required"`
	Rating      float64 `json:"rating" binding:"required"`
}

func ConvertRestToDTO(rest *domain.Restaurant) *RestDTO {

	return &RestDTO{
		ID:          rest.ID,
		Title:       rest.Title,
		Description: rest.Description,
		Address:     rest.Address,
		Rating:      rest.Rating,
	}

}

type ItemDTOAdminRequest struct {
	Title       string  `json:"title" binding:"required"`
	Composition string  `json:"composition" binding:"required"`
	Time        int     `json:"time" binding:"required"`
	Cost        float64 `json:"cost" binding:"required"`
}

type ItemDTOAdminResponse struct {
	ID          int     `json:"id"`
	RestID      int     `json:"rest_id"`
	Title       string  `json:"title"`
	Composition string  `json:"composition"`
	Time        int     `json:"time"`
	Cost        float64 `json:"cost"`
}

func ConvertItemToDTO(item *domain.Item) *ItemDTOAdminResponse {

	return &ItemDTOAdminResponse{
		ID:          item.ID,
		RestID:      item.RestID,
		Title:       item.Title,
		Composition: item.Composition,
		Time:        item.Time,
		Cost:        item.Cost,
	}

}

type ItemDeleteDTO struct {
	RestID int    `json:"rest_id" binding:"required"`
	Title  string `json:"title" binding:"required"`
}

// @Summary Create restaurant
// @Description Создание нового ресторана
// @Tags rest
// @Router /restaurants/create [post]
// @Accept json
// @Produce json
// @Param body body RestDTO true "Название, адрес и рейтинг(опционально описание)"
// @Success 201 {object} RestDTO "Ресторан успешно создан"
// @Failure 400 {object} domain.ErrorResponse "Неправильные входные данные"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *RestaurantHandler) CreateRestaurant(c *gin.Context) {

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	var req RestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error(err.Error())
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	restaurant, err := h.Service.CreateRestaurant(ctx, req.Title, req.Description, req.Address, req.Rating)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTitle) || errors.Is(err, domain.ErrInvalidDescription) || errors.Is(err, domain.ErrInvalidAddress) || errors.Is(err, domain.ErrInvalidRating) || errors.Is(err, domain.ErrWithInsert) {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"CreateRestaurant\" took a lot of time")
	}

	dto := ConvertRestToDTO(restaurant)

	c.JSON(201, dto)

}

// @Summary Restaurants list
// @Description Список ресторанов
// @Tags rest
// @Router /restaurants [get]
// @Produce json
// @Success 200 {array} RestDTO "OK"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
func (h *RestaurantHandler) RestaurantsList(c *gin.Context) {

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	rests, err := h.Service.RestaurantsList(ctx)
	if err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"RestaurantsList\" took a lot of time")
	}

	dtos := []RestDTO{}
	for _, rest := range rests {

		dtos = append(dtos, *ConvertRestToDTO(&rest))

	}

	c.JSON(200, dtos)

}

// @Summary Restaurant menu
// @Description Просмотр меню ресторана
// @Tags rest
// @Router /restaurants/{id}/menu [get]
// @Produce json
// @Param id path int true "ID ресторана"
// @Success 200 {array} ItemDTOAdminResponse "OK"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
func (h *RestaurantHandler) RestaurantMenu(c *gin.Context) {

	strID := c.Param("id")
	id, err := strconv.Atoi(strID)
	if err != nil || id <= 0 {
		slog.Error("Handler \"AddItemsToOrder\" get invalid id")
		c.JSON(400, domain.ErrorResponse{Error: "Get invalid id"})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	menu, err := h.Service.RestaurantMenu(ctx, id)
	if err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn(fmt.Sprintf("\"RestaurantMenu\" from %s title, addresstook a lot of time", strID))
	}

	dtos := []ItemDTOAdminResponse{}
	for _, item := range menu {

		dtos = append(dtos, *ConvertItemToDTO(&item))

	}

	c.JSON(200, dtos)

}

// @Summary Add new items
// @Description Добавление новых блюд в меню
// @Tags rest
// @Router /restaurants/{id} [post]
// @Accept json
// @Param body body []ItemDTOAdminRequest true "Название блюда, состав, время приготовления и цена"
// @Success 201 "Блюда успешно добавлены"
// @Failure 400 {object} domain.ErrorResponse "Неправильные входные данные"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *RestaurantHandler) AddNewItems(c *gin.Context) {

	strID := c.Param("id")
	id, err := strconv.Atoi(strID)
	if err != nil || id <= 0 {
		slog.Error("Handler \"AddItemsToOrder\" get invalid id")
		c.JSON(400, domain.ErrorResponse{Error: "Get invalid id"})
		return
	}

	itemsDTO := []ItemDTOAdminRequest{}
	if err := c.ShouldBindJSON(&itemsDTO); err != nil {
		slog.Error("Handler \"AddNewItems\" get next error when parsing json", "err", err.Error())
		c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	for _, dto := range itemsDTO {

		if err := h.Service.AddNewItems(ctx, id, dto.Title, dto.Composition, dto.Time, dto.Cost); err != nil {
			if errors.Is(err, domain.ErrInvalidTitle) || errors.Is(err, domain.ErrInvalidRestaurantTitle) || errors.Is(err, domain.ErrInvalidComposition) || errors.Is(err, domain.ErrInvalidCost) || errors.Is(err, domain.ErrInvalidTime) || errors.Is(err, domain.ErrWithInsert) {
				c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			} else {
				c.JSON(500, domain.ErrorResponse{Error: err.Error()})
			}
			return
		}

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"AddNewItems\" took a lot of time")
	}

	c.Status(201)

}

// @Summary Delete items
// @Description Удаление блюд из меню
// @Tags rest
// @Router /restaurants/delete [post]
// @Accept json
// @Param body body []ItemDeleteDTO true "ID ресторанов и название блюд"
// @Success 200 "Блюда успешно удалены"
// @Failure 400 {object} domain.ErrorResponse "Неправильные входные данные"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *RestaurantHandler) DeleteItems(c *gin.Context) {

	itemsDTO := []ItemDeleteDTO{}
	if err := c.ShouldBindJSON(&itemsDTO); err != nil {
		slog.Error("Handler \"DeleteItems\" get next error when parsing json", "err", err.Error())
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	for _, dto := range itemsDTO {

		if err := h.Service.DeleteItems(ctx, dto.RestID, dto.Title); err != nil {
			if errors.Is(err, domain.ErrWithDelete) {
				c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			} else {
				c.JSON(500, domain.ErrorResponse{Error: err.Error()})
			}
			return
		}

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"DeleteItems\" took a lot of time")
	}

	c.Status(200)

}
