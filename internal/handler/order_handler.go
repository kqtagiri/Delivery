package handler

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/service"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	Service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {

	return &OrderHandler{
		Service: service,
	}

}

type OrderDTO struct {
	Number  int     `json:"number"`
	Email   string  `json:"email"`
	Address string  `json:"address" binding:"required"`
	Status  string  `json:"status"`
	Time    int     `json:"time"`
	Cost    float64 `json:"cost"`
}

type ItemDTOUser struct {
	Title     string  `json:"title" binding:"required"`
	RestTitle string  `json:"rest_title" binding:"required"`
	Time      int     `json:"time"`
	Cost      float64 `json:"cost"`
}

func ConvertOrderToDTO(order *domain.Order) *OrderDTO {

	return &OrderDTO{
		Number:  order.Number,
		Email:   order.Email,
		Address: order.Address,
		Status:  order.Status,
		Time:    order.Time,
		Cost:    order.Cost,
	}

}

func ConvertItemToDTOUser(item *domain.Item) *ItemDTOUser {

	return &ItemDTOUser{
		Title:     item.Title,
		RestTitle: item.RestTitle,
		Time:      item.Time,
		Cost:      item.Cost,
	}

}

// @Summary Create order
// @Description Создание нового заказа
// @Tags orders
// @Router /orders/create [post]
// @Accept json
// @Produce json
// @Param body body OrderDTO true "Почта и адрес"
// @Success 201 {string} string "Заказ успешно создан"
// @Failure 400 {object} domain.ErrorResponse "Неверные данные"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 404 {object} domain.ErrorResponse "Заказ не найден"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *OrderHandler) CreateOrder(c *gin.Context) {

	slog.Info("Handler started \"CreateOrder\"")

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

	dto := OrderDTO{}
	if err := c.ShouldBindJSON(&dto); err != nil {
		slog.Error(err.Error())
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	order, err := h.Service.CreateOrder(ctx, email, dto.Address)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAddress) || errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrWithInsert) {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		} else if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"CreateOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"CreateOrder\" success")
	c.String(201, "Order created! Number of your order - %s. Use him to add new items.", order.Number)

}

// @Summary Add items to order
// @Description Добавление товаров в заказ
// @Tags orders
// @Router /orders/add/{number} [post]
// @Accept json
// @Produce json
// @Param number path int true "Номер заказа"
// @Param body body []ItemDTOUser true "Название блюда и ресторана"
// @Success 200 "OK"
// @Failure 400 {object} domain.ErrorResponse "Неправильный номер"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 404 {object} domain.ErrorResponse "Заказ или блюдо не найдены"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *OrderHandler) AddItemsToOrder(c *gin.Context) {

	slog.Info("Handler started \"AddItemsToOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil || number <= 0 {
		slog.Error("Handler \"AddItemsToOrder\" get invalid number")
		c.JSON(400, domain.ErrorResponse{Error: "Get invalid number"})
		return
	}

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

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	itemsDTO := []ItemDTOUser{}
	if err := c.ShouldBindJSON(&itemsDTO); err != nil {
		slog.Error("\"AddItemsToOrder\":%w", err)
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	for _, dto := range itemsDTO {

		if err := h.Service.AddItemsToOrder(ctx, number, email, dto.Title, dto.RestTitle); err != nil {
			if errors.Is(err, domain.ErrWithInsert) || errors.Is(err, domain.ErrWithUpdate) {
				c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			} else if errors.Is(err, domain.ErrItemNotFound) || errors.Is(err, domain.ErrOrderNotFound) {
				c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			} else {
				c.JSON(500, domain.ErrorResponse{Error: err.Error()})
			}
			return
		}

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"AddItemsToOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"AddItemsToOrder\" success")
	c.Status(200)

}

// @Summary Delete items from order
// @Description Удаление товаров из заказа
// @Tags orders
// @Router /orders/deleteitems/{number} [post]
// @Accept json
// @Produce json
// @Param number path int true "Номер заказа"
// @Param body body []ItemDTOUser true "Название блюда и ресторана"
// @Success 200 "OK"
// @Failure 400 {object} domain.ErrorResponse "Неправильный номер"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 404 {object} domain.ErrorResponse "Заказ или блюдо не найдены"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *OrderHandler) DeleteItemsFromOrder(c *gin.Context) {

	slog.Info("Handler started \"DeleteItemsFromOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil || number <= 0 {
		slog.Error("Handler \"DeleteItemsFromOrder\" get invalid number")
		c.JSON(400, domain.ErrorResponse{Error: "Get invalid number"})
		return
	}

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

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	itemsDTO := []ItemDTOUser{}
	if err := c.ShouldBindJSON(&itemsDTO); err != nil {
		slog.Error("Handler \"DeleteItemsFromOrder\" get next error when parsing json:%w", err)
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	for _, dto := range itemsDTO {

		if err := h.Service.DeleteItemsFromOrder(ctx, number, email, dto.Title, dto.RestTitle); err != nil {
			if errors.Is(err, domain.ErrWithUpdate) || errors.Is(err, domain.ErrWithDelete) {
				c.JSON(400, domain.ErrorResponse{Error: err.Error()})
			} else if errors.Is(err, domain.ErrItemNotFound) || errors.Is(err, domain.ErrOrderNotFound) {
				c.JSON(404, domain.ErrorResponse{Error: err.Error()})
			} else {
				c.JSON(500, domain.ErrorResponse{Error: err.Error()})
			}
			return
		}

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"DeleteItemsFromOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"DeleteItemsFromOrder\" success")
	c.Status(200)

}

// @Summary Cancel order
// @Description Отмена заказа
// @Tags orders
// @Router /orders/cancel/{number} [patch]
// @Produce json
// @Param number path int true "Номер заказа"
// @Success 200 "OK"
// @Failure 400 {object} domain.ErrorResponse "Неправильный номер"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *OrderHandler) CancelOrder(c *gin.Context) {

	slog.Info("Handler started \"CancelOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil || number <= 0 {
		slog.Error("Haandler \"CancelOrder\"Get invalid number")
		c.JSON(400, domain.ErrorResponse{Error: "Get invalid number"})
		return
	}

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

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	if err := h.Service.CancelOrder(ctx, number, email); err != nil {
		if errors.Is(err, domain.ErrWithUpdate) {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"CancelOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"CancelOrder\" success")
	c.Status(200)

}

// @Summary Get order
// @Description Получение информации о заказе
// @Tags orders
// @Router /orders/{number} [get]
// @Produce json
// @Param number path int true "Номер заказа"
// @Success 200 {object} OrderDTO "OK"
// @Failure 400 {object} domain.ErrorResponse "Неправильный номер"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 404 {object} domain.ErrorResponse "Заказ не найден"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *OrderHandler) GetOrder(c *gin.Context) {

	slog.Info("Handler started \"GetOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil {
		slog.Error("Handler \"GetOrder\" get next error when convert number:%w", err)
		c.JSON(400, domain.ErrorResponse{Error: domain.ErrInvalidNumber.Error()})
		return
	}

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

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	order, err := h.Service.GetOrder(ctx, number, email)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	dto := ConvertOrderToDTO(order)

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"GetOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"GetOrder\" success")
	c.JSON(200, dto)

}

// @Summary Get order details
// @Description Получение подробной информации о заказе
// @Tags orders
// @Router /orders/details/{number} [get]
// @Produce json
// @Param number path int true "Номер заказа"
// @Success 200 {array} ItemDTOUser "OK"
// @Failure 400 {object} domain.ErrorResponse "Неправильный номер"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *OrderHandler) GetOrderDetails(c *gin.Context) {

	slog.Info("Handler started \"GetOrderDetails\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil {
		slog.Error("Handler \"GetOrderDetails\" Get next error when convert number:%w", err)
		c.JSON(400, domain.ErrorResponse{Error: domain.ErrInvalidNumber.Error()})
		return
	}

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

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	items, err := h.Service.GetOrderDetails(ctx, number, email)
	if err != nil {
		c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		return
	}

	dtos := []ItemDTOUser{}
	for _, item := range *items {

		dtos = append(dtos, *ConvertItemToDTOUser(&item))

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"GetOrderDetails\" took a lot of time")
	}

	slog.Info("Handler ended \"GetOrderDetails\" success")
	c.JSON(200, dtos)

}

// @Summary Confirm order
// @Description Подтверждение заказа
// @Tags orders
// @Router /orders/confirm/{number} [post]
// @Accept json
// @Produce json
// @Param number path int true "Номер заказа"
// @Success 200 "OK"
// @Failure 400 {object} domain.ErrorResponse "Неправильный номер"
// @Failure 401 {object} domain.ErrorResponse "Почта не найдена"
// @Failure 404 {object} domain.ErrorResponse "Заказ или пользователь не найден"
// @Failure 500 {object} domain.ErrorResponse "Произошла ошибка на сервере"
// @Security Auth
func (h *OrderHandler) ConfirmOrder(c *gin.Context) {

	slog.Info("Handler started \"ConfirmOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil {
		slog.Error("Handler \"ConfirmOrder\" Get next error when convert number:%w", err)
		c.JSON(400, domain.ErrorResponse{Error: "Get invalid number"})
		return
	}

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

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	if err := h.Service.ConfirmOrder(ctx, number, email); err != nil {
		if errors.Is(err, domain.ErrWithUpdate) || errors.Is(err, domain.ErrStatusNotCreated) {
			c.JSON(400, domain.ErrorResponse{Error: err.Error()})
		} else if errors.Is(err, domain.ErrOrderNotFound) || errors.Is(err, domain.ErrUserNotFound) {
			c.JSON(404, domain.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(500, domain.ErrorResponse{Error: err.Error()})
		}
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"ConfirmOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"ConfirmOrder\" success")
	c.Status(200)

}
