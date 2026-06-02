package main

import (
	"context"
	"delivery/internal/database"
	"delivery/internal/domain"
	"delivery/internal/handler"
	"delivery/internal/repository"
	"delivery/internal/service"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(role string) gin.HandlerFunc {

	return func(c *gin.Context) {

		str := c.GetHeader("Authorization")
		if str == "" {
			slog.Error("Get invalid auth header")
			c.JSON(400, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(str, "Bearer ")
		if tokenStr == "" {
			slog.Error("Get invalid auth header")
			c.JSON(400, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}

		claims := &domain.AccessTokenClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Invalid singing method:%v", token.Header["alg"])
			}
			return []byte(os.Getenv("ACCESS_SECRET_KEY")), nil
		})

		if err != nil {
			slog.Error("AuthMiddleware get next error:", err)
			c.JSON(500, gin.H{"error": err})
			c.Abort()
			return
		}

		if token.Valid {
			if claims.Role != role {
				slog.Error("Invalid role")
				c.JSON(403, gin.H{"error": "Invalid role"})
				c.Abort()
				return
			}
			c.Set("user_id", claims.UserId)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Next()
		} else {
			slog.Error("Invalid Token")
			c.JSON(400, gin.H{"error": "Invalid token"})
			c.Abort()
		}

	}

}

func main() {

	ctx := context.Context(context.Background())
	db, err := database.NewDB(ctx)
	if err != nil {
		slog.Error("Get next err when connect to database:", err)
		return
	}

	addr := os.Getenv("REDIS_HOST")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	user_repo := repository.NewUserRepositoryStruct(db)
	user_service := service.NewUserService(user_repo)
	user_handler := handler.NewUserHandler(user_service)

	rest_repo := repository.NewRestaurantRepositoryStruct(db, rdb)
	rest_service := service.NewRestaurantService(rest_repo)
	rest_handler := handler.NewRestaurantHandler(rest_service)

	order_repo := repository.NewOrderRepositoryStruct(db)
	order_service := service.NewOrderService(order_repo)
	order_handler := handler.NewOrderHandler(order_service)

	if err := user_repo.CreateUserTable(); err != nil {
		slog.Error("Get next error when create users table:", err)
		return
	}

	if err := rest_repo.CreateMenuTable(); err != nil {
		slog.Error("Get next error when create menu table:", err)
		return
	}

	if err := rest_repo.CreateRestaurantTable(); err != nil {
		slog.Error("Get next error when create restaurants table:", err)
		return
	}

	if err := order_repo.CreateItemsInOrdersTable(); err != nil {
		slog.Error("Get next error when create orders_items table:", err)
		return
	}

	if err := order_repo.CreateOrderTable(); err != nil {
		slog.Error("Get next error when create orders table:", err)
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
	rest.POST("/create", AuthMiddleware(domain.AdminRole), rest_handler.CreateRestaurant) //admin
	rest.GET("", rest_handler.RestaurantsList)
	rest.GET("/:title", rest_handler.RestaurantMenu)
	rest.POST("/:title", AuthMiddleware(domain.AdminRole), rest_handler.AddNewItems)   //admin
	rest.DELETE("/:title", AuthMiddleware(domain.AdminRole), rest_handler.DeleteItems) //admin

	users := r.Group("/users")
	users.POST("/register", user_handler.RegisterAccount)
	users.GET("/login", user_handler.Authorization)
	users.GET("/refresh", user_handler.Refresh)
	users.PATCH("/replenish", AuthMiddleware(domain.UserRole), user_handler.ReplenishBalance) // user
	users.GET("", AuthMiddleware(domain.UserRole), user_handler.UserInfo)                     //user
	users.GET("/all", AuthMiddleware(domain.AdminRole), user_handler.AllUsersInfo)            //admin
	//users.GET("/history/:email", user_handler.UserHistoryInfo)

	order := r.Group("/orders")
	order.Use(AuthMiddleware(domain.UserRole))
	order.POST("/create", order_handler.CreateOrder)                         //user
	order.POST("/add/:number", order_handler.AddItemsToOrder)                //user
	order.DELETE("/deleteitems/:number", order_handler.DeleteItemsFromOrder) //user
	order.DELETE("/delete/:number", order_handler.CancelOrder)               //user
	order.GET("/:number", order_handler.GetOrder)                            //user
	order.GET("/details/:number", order_handler.GetOrderDetails)             //user
	order.POST("/confirm/:number", order_handler.ConfirmOrder)               //user

	server := http.Server{
		Addr:    ":9111",
		Handler: r,
	}

	go func() {

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(err.Error())
			return
		}

	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	context, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := server.Shutdown(context); err != nil {
		slog.Error(err.Error())
		return
	}

	if err := db.Close(); err != nil {
		slog.Error(err.Error())
		return
	}

}
