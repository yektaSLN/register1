package routes

import (
	"login/config"
	"login/handler"
	"login/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config, authHandler *handler.AuthHandler, productHandler *handler.ProductHandler, rateLimiter *middleware.RateLimiter) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	api := router.Group("/api")

	//auth := api.Group("/auth")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", rateLimiter.Middleware(), authHandler.Login)
		api.POST("/refresh", authHandler.RefreshToken) //*
		api.POST("/forgot-password", authHandler.ForgotPassword)
		api.POST("/reset-password", authHandler.ResetPassword)
	}

	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		users.GET("/me", authHandler.GetMe)
	}

	products := api.Group("/products")
	products.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		products.POST("", productHandler.Create)
		products.GET("", productHandler.GetAll)
		products.GET("/:id", productHandler.GetByID)
		products.PUT("/:id", productHandler.Update)
		products.DELETE("/:id", productHandler.Delete)
	}

	return router
}
