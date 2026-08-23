package main

import (
	"log"

	"login/config"
	"login/database"
	"login/handler"
	"login/kafka"
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

	//redis
	redisClient := redis.NewClient(cfg.RedisAddr)

	//kafka
	kafkaProducer := kafka.NewProducer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
	)
	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Println("failed to close kafka producer:", err)
		}
	}()

	//rate limiters
	loginRateLimiter := middleware.NewRateLimiter(
		redisClient,
		"login",
		5,
		60,
	)

	registerRateLimiter := middleware.NewRateLimiter(
		redisClient,
		"register",
		5,
		60,
	)

	refreshRateLimiter := middleware.NewRateLimiter(
		redisClient,
		"refresh",
		10,
		60,
	)

	forgotPasswordRateLimiter := middleware.NewRateLimiter(
		redisClient,
		"forgot-password",
		3,
		60,
	)

	resetPasswordRateLimiter := middleware.NewRateLimiter(
		redisClient,
		"reset-password",
		5,
		60,
	)

	productRateLimiter := middleware.NewRateLimiter(
		redisClient,
		"product",
		5,
		60,
	)

	//login database
	loginDB, err := database.Connect(cfg, cfg.DBName)
	if err != nil {
		log.Fatal("failed to connect to login database:", err)
	}

	if err := loginDB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("failed to migrate login database:", err)
	}

	userRepository := repository.NewUserRepository(loginDB)

	authService := service.NewAuthService(
		userRepository,
		cfg.JWTSecret,
		cfg.JWTExpiration,
		redisClient,
		kafkaProducer,
	)

	authHandler := handler.NewAuthHandler(authService)

	productDB, err := database.Connect(cfg, cfg.ProductDBName)
	if err != nil {
		log.Fatal("failed to connect to products database:", err)
	}

	if err := productDB.AutoMigrate(&models.Product{}); err != nil {
		log.Fatal("failed to migrate products database:", err)
	}

	productRepository := repository.NewProductRepository(productDB)

	productService := service.NewProductService(
		productRepository,
		kafkaProducer,
	)

	productHandler := handler.NewProductHandler(productService)

	router := routes.SetupRouter(
		cfg,
		authHandler,
		productHandler,
		loginRateLimiter,
		registerRateLimiter,
		refreshRateLimiter,
		forgotPasswordRateLimiter,
		resetPasswordRateLimiter,
		productRateLimiter,
	)

	log.Println("server is running on port", cfg.ServerPort)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
