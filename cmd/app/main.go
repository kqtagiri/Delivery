package main

import (
	"context"
	"delivery/internal/database"
	"delivery/internal/handler"
	"delivery/internal/repository"
	"delivery/internal/service"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func main() {

	ctx := context.Context(context.Background())
	db, err := database.NewDB(ctx)
	if err != nil {
		slog.Error("Get next err when connect to database:%w", err)
		return
	}

	user_repo := repository.NewUserRepositoryStruct(db)
	user_service := service.NewUserService(user_repo)
	user_handler := handler.NewUserHandler(user_service)

	rest_repo := repository.NewRestaurantRepositoryStruct(db)
	rest_service := service.NewRestaurantService(rest_repo)
	rest_handler := handler.NewRestaurantHandler(rest_service)

	order_repo := repository.NewOrderRepositoryStruct(db)
	order_service := service.NewOrderService(order_repo)
	order_handler := handler.NewOrderHandler(order_service)

	if err := user_repo.CreateUserTable(); err != nil {
		slog.Error("Get next error when create users table:%w", err)
		return
	}

	if err := rest_repo.CreateMenuTable(); err != nil {
		slog.Error("Get next error when create menu table:%w", err)
		return
	}

	if err := rest_repo.CreateRestaurantTable(); err != nil {
		slog.Error("Get next error when create restaurants table:%w", err)
		return
	}

	if err := order_repo.CreateItemsInOrdersTable(); err != nil {
		slog.Error("Get next error when create orders_items table:%w", err)
		return
	}

	if err := order_repo.CreateOrderTable(); err != nil {
		slog.Error("Get next error when create orders table:%w", err)
		return
	}

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"Method":  c.Request.Method,
			"Status":  200,
			"Message": "Server is open!",
		})
	})

	rest := r.Group("/restaurants")
	rest.POST("/create", rest_handler.CreateRestaurant)
	rest.GET("", rest_handler.RestaurantsList)
	rest.GET("/:title", rest_handler.RestaurantMenu)
	rest.POST("/:title", rest_handler.AddNewItems)
	rest.DELETE("/:title", rest_handler.DeleteItems)

	users := r.Group("/users")
	users.POST("/register", user_handler.RegisterAccount)
	users.PATCH("/replenish/:email", user_handler.ReplenishBalance)
	users.GET("/:email", user_handler.UserInfo)
	users.GET("", user_handler.AllUsersInfo)
	//users.GET("/history/:email", user_handler.UserHistoryInfo)

	order := r.Group("/orders")
	order.POST("/create", order_handler.CreateOrder)
	order.POST("/add/:number", order_handler.AddItemsToOrder)
	order.DELETE("/deleteitems/:number", order_handler.DeleteItemsFromOrder)
	order.DELETE("/delete/:number", order_handler.CancelOrder)
	order.GET("/:number", order_handler.GetOrder)
	order.GET("/details/:number", order_handler.GetOrderDetails)
	order.POST("/confirm/:number", order_handler.ConfirmOrder)

	r.Run(":9111")

}
