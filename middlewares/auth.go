package middleware

import (
	"net/http"
	"strings"

	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "Se requiere token de autorización", nil)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.Error(c, http.StatusUnauthorized, "Formato de token inválido", nil)
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "Token inválido o expirado", nil)
			c.Abort()
			return
		}

		// Inyectamos datos críticos en el contexto
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role) // Cambiado de RoleID a Role (string) según nuestro modelo User

		c.Next()
	}
}
