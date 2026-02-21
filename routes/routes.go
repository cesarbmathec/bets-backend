package routes

import (
	"os"
	"strings"

	"github.com/cesarbmathec/bets-backend/controllers"
	middleware "github.com/cesarbmathec/bets-backend/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	// --- MIDDLEWARES GLOBALES ---
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())

	// Configuración de CORS Profesional
	allowedOrigins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: os.Getenv("CORS_ALLOW_CREDENTIALS") == "true",
	}))

	r.Use(middleware.SecurityHeaders())

	// Documentación API (Swagger)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(1),
		ginSwagger.PersistAuthorization(true),
	))

	// --- RUTAS DE LA API V1 ---
	api := r.Group("/api/v1")
	{
		// Limiter específico para Auth
		authLimiter := middleware.NewRateLimiter()

		// --- RUTAS PÚBLICAS (AUTH) --- //
		auth := api.Group("/auth")
		{
			auth.POST("/login", authLimiter.Middleware(), controllers.Login)
			auth.POST("/register", authLimiter.Middleware(), controllers.Register)
		}

		// --- RUTAS PÚBLICAS DE CONSULTA --- //
		tournaments := api.Group("/tournaments")
		{
			tournaments.GET("/", controllers.GetTournaments)
			tournaments.GET("/id/:id", controllers.GetTournamentByID)
			tournaments.GET("/id/:id/leaderboard", controllers.GetTournamentLeaderboard)
			tournaments.GET("/s/:slug", controllers.GetTournamentBySlug)
			tournaments.GET("/id/:id/events", controllers.GetTournamentEvents)
			tournaments.GET("/id/:id/sessions", controllers.GetTournamentSessions)
		}

		sessionEvents := api.Group("/session-events")
		{
			sessionEvents.GET("/:id", controllers.GetSessionByID)
		}

		events := api.Group("/events")
		{
			events.GET("/s/:slug", controllers.GetEventBySlug)
			events.GET("/id/:id", controllers.GetEventByID)
			events.GET("/id/:id/selections", controllers.GetEventSelections)
		}

		// --- RUTAS PROTEGIDAS (Usuario autenticado) --- //
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Perfil del usuario autenticado
			protected.GET("/me", controllers.GetMyProfile)

			// --- RUTAS DE USUARIOS COMUNES --- //
			userRoutes := protected.Group("")
			{
				// Inscripción y picks
				userRoutes.POST("/tournaments/:id/join", controllers.JoinTournament)
				userRoutes.POST("/tournaments/:id/sessions/picks", controllers.SubmitPicksBySession)

				// Billetera
				userRoutes.GET("/wallet/balance", controllers.GetBalance)
				userRoutes.POST("/wallet/deposit", controllers.DepositMoney)
				userRoutes.GET("/wallet/history", controllers.GetTransactionHistory)
				userRoutes.GET("/wallet/statistics", controllers.GetUserStatistics)

				// Ver mis picks de sesión
				userRoutes.GET("/my-sessions/:session_id/picks", controllers.GetSessionPicks)

				// Métodos de pago (retiro)
				userRoutes.GET("/payment-methods", controllers.GetPaymentMethods)
				userRoutes.POST("/payment-methods", controllers.CreatePaymentMethod)
				userRoutes.DELETE("/payment-methods/:id", controllers.DeletePaymentMethod)
			}
		}

		// --- RUTAS PROTEGIDAS DE ADMINISTRADOR --- //
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.RequireAdmin())
		{
			// Gestión de usuarios
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("/", controllers.GetUsers)
				adminUsers.GET("/:id", controllers.GetUserByID)
				adminUsers.PATCH("/:id/role", controllers.UpdateUserRole)
				adminUsers.PATCH("/:id/status", controllers.UpdateUserStatus)
			}

			// Gestión de Torneos
			adminTournaments := admin.Group("/tournaments")
			{
				adminTournaments.POST("/", controllers.CreateTournament)
				adminTournaments.PATCH("/:id/status", controllers.UpdateTournamentStatus)
			}

			// Gestión de Sesiones
			adminSessions := admin.Group("/sessions")
			{
				adminSessions.POST("/", controllers.CreateSession)
				adminSessions.PATCH("/:id/status", controllers.UpdateSessionStatus)
			}

			// Gestión de Eventos
			adminEvents := admin.Group("/events")
			{
				adminEvents.POST("/", controllers.CreateEvent)
				adminEvents.POST("/selections", controllers.CreateSelection)
				adminEvents.POST("/id/:id/competitors", controllers.SetEventCompetitors)
				adminEvents.POST("/:id/settle", controllers.SettleEvent)
			}
		}
	}

	return r
}
