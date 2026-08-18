package main

import (
	"log"

	"login/config"
	"login/database"
	"login/handler"
	"login/middleware"
	"login/models"
	"login/redis"
	"login/repository"
	"login/routes"
	"login/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	//adding redis
	redisClient := redis.NewClient(cfg.RedisAddr)

	loginRateLimiter := middleware.NewRateLimiter(redisClient, "login", 5, 60)

	registerRateLimiter := middleware.NewRateLimiter(redisClient, "register", 5, 60)

	refreshRateLimiter := middleware.NewRateLimiter(redisClient, "refresh", 10, 60)

	forgotPasswordRateLimiter := middleware.NewRateLimiter(redisClient, "forgot-password", 3, 60)

	resetPasswordRateLimiter := middleware.NewRateLimiter(redisClient, "reset-password", 5, 60)
	productRateLimiter := middleware.NewRateLimiter(redisClient, "product", 5, 60)

	loginDB, err := database.Connect(cfg, cfg.DBName)

	if err != nil {
		log.Fatal("failed to connect to login database:", err)
	}

	err = loginDB.AutoMigrate(&models.User{})

	if err != nil {
		log.Fatal("failed to migrate login database:", err)
	}

	userRepository := repository.NewUserRepository(
		loginDB,
	)

	authService := service.NewAuthService(
		userRepository,
		cfg.JWTSecret,
		cfg.JWTExpiration,

		redisClient, //*
	)

	authHandler := handler.NewAuthHandler(authService)

	productDB, err := database.Connect(cfg, cfg.ProductDBName)

	if err != nil {
		log.Fatal("failed to connect to products database:", err)
	}

	err = productDB.AutoMigrate(&models.Product{})

	if err != nil {
		log.Fatal("failed to migrate products database:", err)
	}

	productRepository := repository.NewProductRepository(productDB)
	productService := service.NewProductService(productRepository)
	productHandler := handler.NewProductHandler(productService)

	router := routes.SetupRouter(cfg, authHandler, productHandler, loginRateLimiter, registerRateLimiter, refreshRateLimiter, forgotPasswordRateLimiter, resetPasswordRateLimiter, productRateLimiter)

	log.Println("server is running on port", cfg.ServerPort)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
