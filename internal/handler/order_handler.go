package handler

import (
	"context"
	"delivery/internal/domain"
	"delivery/internal/service"
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
	Email   string  `json:"email" binding:"email,required"`
	Address string  `json:"address" binding:"required"`
	Status  string  `json:"status"`
	Time    int     `json:"time"`
	Cost    float64 `json:"cost"`
}

type ItemDTOUser struct {
	Title     string  `json:"title" binding:"required"`
	RestTitle string  `json:"restaurant_title" binding:"required"`
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

func (h *OrderHandler) CreateOrder(c *gin.Context) {

	slog.Info("Handler started \"CreateOrder\"")

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	dto := OrderDTO{}
	if err := c.ShouldBindJSON(&dto); err != nil {
		slog.Error(err.Error())
		c.JSON(400, gin.H{"error": err})
		return
	}

	err, order := h.Service.CreateOrder(&ctx, dto.Email, dto.Address)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"CreateOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"CreateOrder\" success")
	c.String(201, "Order created! Number of your order - %s. Use him to add new items.", order.Number)

}

func (h *OrderHandler) AddItemsToOrder(c *gin.Context) {

	slog.Info("Handler started \"AddItemsToOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil || number <= 0 {
		slog.Error("Handler \"AddItemsToOrder\" get invalid number")
		c.JSON(400, gin.H{"error": "Get invalid number"})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	itemsDTO := []ItemDTOUser{}
	if err := c.ShouldBindJSON(&itemsDTO); err != nil {
		slog.Error("\"AddItemsToOrder\":%w", err)
		c.JSON(500, gin.H{"error": err})
		return
	}

	for _, dto := range itemsDTO {

		if err := h.Service.AddItemsToOrder(&ctx, number, dto.Title, dto.RestTitle); err != nil {
			c.JSON(500, gin.H{"error": err})
			return
		}

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"AddItemsToOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"AddItemsToOrder\" success")
	c.Status(200)

}

func (h *OrderHandler) DeleteItemsFromOrder(c *gin.Context) {

	slog.Info("Handler started \"DeleteItemsFromOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil || number <= 0 {
		slog.Error("Handler \"DeleteItemsFromOrder\" get invalid number")
		c.JSON(400, gin.H{"error": "Get invalid number"})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	itemsDTO := []ItemDTOUser{}
	if err := c.ShouldBindJSON(&itemsDTO); err != nil {
		slog.Error("Handler \"DeleteItemsFromOrder\" get next error when parsing json:%w", err)
		c.JSON(500, gin.H{"error": err})
		return
	}

	for _, dto := range itemsDTO {

		if err := h.Service.DeleteItemsFromOrder(&ctx, number, dto.Title, dto.RestTitle); err != nil {
			c.JSON(400, gin.H{"error": err})
			return
		}

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"DeleteItemsFromOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"DeleteItemsFromOrder\" success")
	c.Status(200)

}

func (h *OrderHandler) DeleteOrder(c *gin.Context) {

	slog.Info("Handler started \"DeleteOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil || number <= 0 {
		slog.Error("Haandler \"DeleteOrder\"Get invalid number")
		c.JSON(400, gin.H{"error": "Get invalid number"})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	if err := h.Service.DeleteOrder(&ctx, number); err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"DeleteOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"DeleteOrder\" success")
	c.Status(200)

}

func (h *OrderHandler) OrderInfo(c *gin.Context) {

	slog.Info("Handler started \"OrderInfo\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil {
		slog.Error("Handler \"OrderInfo\" Get next error when convert number:%w", err)
		c.JSON(400, gin.H{"error": "Get invalid number"})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	err, order := h.Service.OrderInfo(&ctx, number)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	dto := ConvertOrderToDTO(order)

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"OrderInfo\" took a lot of time")
	}

	slog.Info("Handler ended \"OrderInfo\" success")
	c.JSON(200, dto)

}

func (h *OrderHandler) OrderDetailInfo(c *gin.Context) {

	slog.Info("Handler started \"OrderDetailInfo\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil {
		slog.Error("Handler \"OrderDetailInfo\" Get next error when convert number:%w", err)
		c.JSON(400, gin.H{"error": "Get invalid number"})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	err, items := h.Service.OrderDetailInfo(&ctx, number)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	dtos := []ItemDTOUser{}
	for _, item := range *items {

		dtos = append(dtos, *ConvertItemToDTOUser(&item))

	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"OrderDetailInfo\" took a lot of time")
	}

	slog.Info("Handler ended \"OrderDetailInfo\" success")
	c.JSON(200, dtos)

}

func (h *OrderHandler) ConfirmOrder(c *gin.Context) {

	slog.Info("Handler started \"ConfirmOrder\"")

	number_string := c.Param("number")
	number, err := strconv.Atoi(number_string)
	if err != nil {
		slog.Error("Handler \"ConfirmOrder\" Get next error when convert number:%w", err)
		c.JSON(400, gin.H{"error": "Get invalid number"})
		return
	}
	email, check := c.GetQuery("email")
	if !check {
		slog.Error("User didn`t write his email!")
		c.JSON(400, gin.H{"error": "You didn`t write your email!"})
		return
	}

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	start := time.Now()

	if err := h.Service.ConfirmOrder(&ctx, number, email); err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if time.Since(start) > 4*time.Second {
		slog.Warn("\"ConfirmOrder\" took a lot of time")
	}

	slog.Info("Handler ended \"ConfirmOrder\" success")
	c.Status(200)

}
