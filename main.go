package main

import (
	"os"

	"login/config"
	"login/database"
	"login/handler"
	"login/kafka"
	"login/logger"
	"login/middleware"
	"login/models"
	"login/redis"
	"login/repository"
	"login/routes"
	"login/service"

	"github.com/rs/zerolog"
)

func main() {

	logger.Configure()

	log := logger.New(os.Stdout)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to load config")
	}

	redisClient := redis.NewClient(
		cfg.RedisAddr,
	)

	kafkaProducer := kafka.NewProducer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
	)

	defer func() {
		if err := kafkaProducer.Close(); err != nil {
			log.Error().
				Err(err).
				Msg("failed to close kafka producer")
		}
	}()

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

	loginDB, err := database.Connect(
		cfg,
		cfg.DBName,
	)

	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to connect to login database")
	}

	if err := loginDB.AutoMigrate(
		&models.User{},
	); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to migrate login database")
	}

	userRepository := repository.NewUserRepository(
		loginDB,
	)

	authService := service.NewAuthService(
		userRepository,
		cfg.JWTSecret,
		cfg.JWTExpiration,
		redisClient,
		kafkaProducer,
	)

	authHandler := handler.NewAuthHandler(
		authService,
	)

	productDB, err := database.Connect(
		cfg,
		cfg.ProductDBName,
	)

	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to connect to products database")
	}

	if err := productDB.AutoMigrate(
		&models.Product{},
	); err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to migrate products database")
	}

	productRepository := repository.NewProductRepository(
		productDB,
	)

	productService := service.NewProductService(
		productRepository,
		kafkaProducer,
	)

	productHandler := handler.NewProductHandler(
		productService,
	)

	router := routes.SetupRouter(
		cfg,
		authHandler,
		productHandler,
		kafkaProducer,
		loginRateLimiter,
		registerRateLimiter,
		refreshRateLimiter,
		forgotPasswordRateLimiter,
		resetPasswordRateLimiter,
		productRateLimiter,
	)

	log.Info().
		Str("port", cfg.ServerPort).
		Msg("server is running")

	zerolog.DefaultContextLogger = &log

	if err := router.Run(
		":" + cfg.ServerPort,
	); err != nil {

		log.Fatal().
			Err(err).
			Msg("failed to start server")
	}
}
