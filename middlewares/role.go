package middleware

import (
	"net/http"

	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
)

// RequireAdmin middleware verifica que el usuario tenga rol de administrador
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el rol del contexto (establecido por AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			utils.Error(c, http.StatusUnauthorized, "No autorizado", nil)
			c.Abort()
			return
		}

		// Verificar que el rol sea admin
		if role != "admin" {
			utils.Error(c, http.StatusForbidden, "Acceso denegado. Se requiere rol de administrador", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireUser middleware verifica que el usuario esté autenticado
// (ya existe en AuthMiddleware, pero es útil para documentación)
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// El AuthMiddleware ya verifica que el token sea válido
		// Si llega aquí, el usuario está autenticado
		c.Next()
	}
}
