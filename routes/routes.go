package routes

import (
	"login/config"
	"login/handler"
	"login/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	productHandler *handler.ProductHandler,
	loginRateLimiter *middleware.RateLimiter,
	registerRatelimiter *middleware.RateLimiter,
	refreshRateLimiter *middleware.RateLimiter,
	forgotPasswordRateLimiter *middleware.RateLimiter,
	resetPasswordRateLimiter *middleware.RateLimiter,
	productRateLimiter *middleware.RateLimiter) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	api := router.Group("/api")

	//auth := api.Group("/auth")
	{
		api.POST("/register", registerRatelimiter.Middleware(), authHandler.Register)
		api.POST("/login", loginRateLimiter.Middleware(), authHandler.Login)
		api.POST("/refresh", refreshRateLimiter.Middleware(), authHandler.RefreshToken) //*
		api.POST("/forgot-password", forgotPasswordRateLimiter.Middleware(), authHandler.ForgotPassword)
		api.POST("/reset-password", resetPasswordRateLimiter.Middleware(), authHandler.ResetPassword)
	}

	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		users.GET("/me", authHandler.GetMe)
	}

	products := api.Group("/products")
	products.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		products.POST("", productRateLimiter.Middleware(), productHandler.Create)
		products.GET("", productHandler.GetAll)
		products.GET("/export", productHandler.Export)
		products.GET("/:id", productHandler.GetByID)
		products.PUT("/:id", productHandler.Update)
		products.DELETE("/:id", productHandler.Delete)

	}

	return router
}
