package routes

import (
	"login/config"
	"login/handler"
	"login/middleware"

	"github.com/gin-gonic/gin"
)

//getting the setting and handler methods and returning the main router of gin

func SetupRouter(cfg *config.Config, authHandler *handler.AuthHandler) *gin.Engine {

	//release mode is set for production and it dowsnt show extra debug messages
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	//logger used for debugging and recovery for server stability
	router.Use(gin.Logger(), gin.Recovery())

	api := router.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	users := api.Group("/users")
	users.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		users.GET("/me", authHandler.GetMe)
	}

	return router
}
