package middleware

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
)

type clientLimiter struct {
	tokens     float64
	lastRefill time.Time
}

type limiter struct {
	mu       sync.Mutex
	limiters map[string]*clientLimiter
	limit    float64
	burst    int
}

func NewRateLimiter() *limiter {
	// Configuraciones desde .env o valores por defecto
	return &limiter{
		limiters: make(map[string]*clientLimiter),
		limit:    120.0 / 60.0, // 120 peticiones por minuto
		burst:    30,
	}
}

func (l *limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("RATE_LIMIT_ENABLED") == "false" {
			c.Next()
			return
		}

		key := c.ClientIP()
		l.mu.Lock()
		cl, ok := l.limiters[key]
		if !ok {
			cl = &clientLimiter{tokens: float64(l.burst), lastRefill: time.Now()}
			l.limiters[key] = cl
		}

		now := time.Now()
		elapsed := now.Sub(cl.lastRefill).Seconds()
		cl.tokens += elapsed * l.limit
		if cl.tokens > float64(l.burst) {
			cl.tokens = float64(l.burst)
		}
		cl.lastRefill = now

		if cl.tokens < 1 {
			l.mu.Unlock()
			utils.Error(c, http.StatusTooManyRequests, "Demasiadas solicitudes. Intente más tarde.", nil)
			c.Abort()
			return
		}

		cl.tokens -= 1
		l.mu.Unlock()
		c.Next()
	}
}
