package middleware

import (
	"strings"

	"login/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

//this class is only for logged in users

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(401, utils.NewErrorResponse(utils.ErrInvalidToken))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			c.JSON(401, utils.NewErrorResponse(utils.ErrInvalidToken))
			c.Abort()
			return
		}

		token, err := utils.ValidateToken(parts[1], jwtSecret)

		if err != nil || !token.Valid {
			c.JSON(401, utils.NewErrorResponse(utils.ErrInvalidToken))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(401, utils.NewErrorResponse(utils.ErrInvalidToken))
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(401, utils.NewErrorResponse(utils.ErrInvalidToken))
			c.Abort()
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(401, utils.NewErrorResponse(utils.ErrInvalidToken))
			c.Abort()
			return
		}

		c.Set("user_id", uint(userID))
		c.Set("username", username)

		c.Next()
	}
}
