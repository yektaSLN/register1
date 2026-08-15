package main

import (
	"log"

	"login/config"
	"login/database"
	"login/handler"
	"login/models"
	"login/repository"
	"login/routes"
	"login/service"
)

func main() {

	//with this config we want to read the env file
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	userRepository := repository.NewUserRepository(db)

	authService := service.NewAuthService(userRepository, cfg.JWTSecret, cfg.JWTExpiration)

	authHandler := handler.NewAuthHandler(authService)

	router := routes.SetupRouter(cfg, authHandler)

	log.Println("server is running on port", cfg.ServerPort)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
