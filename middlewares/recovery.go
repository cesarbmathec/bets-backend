package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Loguear el stack trace para debugging
				debug.PrintStack()
				utils.Error(c, http.StatusInternalServerError, "Error crítico del sistema", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
