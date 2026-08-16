package main

import (
	"log"

	"login/config"
	"login/database"
	"login/handler"
	"login/models"
	"login/redis" // *
	"login/repository"
	"login/routes"
	"login/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	redisClient := redis.NewClient(cfg.RedisAddr) // *

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
		redisClient, // *
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

	router := routes.SetupRouter(cfg, authHandler, productHandler)

	log.Println("server is running on port", cfg.ServerPort)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
