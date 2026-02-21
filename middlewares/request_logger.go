package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger formatea los logs de cada petición para auditoría técnica
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		path := param.Path
		if path == "" && param.Request != nil {
			path = param.Request.URL.Path
		}

		// Extraemos el Origin para monitorear desde dónde vienen las apuestas
		origin := param.Request.Header.Get("Origin")
		if origin == "" {
			origin = "Direct/Server-Side"
		}

		// Formato profesional: IP - Tiempo "METODO RUTA" STATUS LATENCIA | Origin
		return fmt.Sprintf(
			"[API] %s - [%s] \"%s %s\" %d %s | Origin: %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			path,
			param.StatusCode,
			param.Latency,
			origin,
		)
	})
}
