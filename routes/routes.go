package routes

import (
	"fmt"
	"log"
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

	// Deshabilitar redirect de trailing slash para evitar problemas de CORS
	r.RedirectTrailingSlash = false

	// --- MIDDLEWARES GLOBALES ---
	r.Use(middleware.RequestLogger())
	r.Use(middleware.Recovery())

	served := os.Getenv("GIN_MODE") == "release"

	// Configuración de CORS Profesional
	allowedOrigins := parseCSVEnv("CORS_ALLOWED_ORIGINS")

	// En desarrollo, permitir todos los orígenes para evitar problemas de CORS
	if !served {
		allowedOrigins = []string{"*"}
	}

	if len(allowedOrigins) == 0 && served {
		log.Fatal("CORS_ALLOWED_ORIGINS requerido en GIN_MODE=release")
	}

	allowCredentials := true
	if strings.EqualFold(os.Getenv("CORS_ALLOW_CREDENTIALS"), "false") {
		allowCredentials = false
	}

	// Si se permite *, no se pueden usar credenciales
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowCredentials = false
			break
		}
	}

	fmt.Printf("Logger AllowCredentials: %v\n", allowCredentials)
	fmt.Printf("Logger AllowedOrigins: %v\n", allowedOrigins)

	// Configuración CORS - permitir todos los orígenes en desarrollo
	if !served {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
			ExposeHeaders:    []string{"Content-Length", "Authorization"},
			AllowCredentials: false,
			MaxAge:           86400,
		}))
	} else {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     allowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
			ExposeHeaders:    []string{"Content-Length", "Authorization"},
			AllowCredentials: allowCredentials,
			MaxAge:           86400,
		}))
	}

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
		categoryCtrl := controllers.NewCategoryController()
		categories := api.Group("/categories")
		{
			// Ruta sin slash para evitar redirección 301
			categories.GET("", categoryCtrl.GetCategories)
			categories.GET("/", categoryCtrl.GetCategories)
			categories.GET("/:id", categoryCtrl.GetCategoryByID)
		}

		tournaments := api.Group("/tournaments")
		{
			// Ruta sin slash para evitar redirección 301
			tournaments.GET("", controllers.GetTournaments)
			tournaments.GET("/", controllers.GetTournaments)
			tournaments.GET("/id/:id", controllers.GetTournamentByID)
			tournaments.GET("/id/:id/leaderboard", controllers.GetTournamentLeaderboard)
			tournaments.GET("/s/:slug", controllers.GetTournamentBySlug)
			tournaments.GET("/id/:id/events", controllers.GetTournamentEvents)
			tournaments.GET("/id/:id/sessions", controllers.GetTournamentSessions)
			tournaments.GET("/my-tournaments", controllers.GetMyTournaments)
		}

		sessionEvents := api.Group("/session-events")
		{
			sessionEvents.GET("/:id", controllers.GetSessionByID)
			sessionEvents.GET("/:id/events", controllers.GetSessionEvents)
		}

		events := api.Group("/events")
		{
			events.GET("/", controllers.GetGlobalEvents)
			events.GET("/:id", controllers.GetEventByID)
			events.GET("/:id/selections", controllers.GetEventSelections)
			events.GET("/tournament/:tournament_id", controllers.GetTournamentEventsByTournament)
		}

		// --- RUTAS DE COMPETIDORES (PÚBLICO) --- //
		competitors := api.Group("/competitors")
		{
			competitors.GET("", controllers.GetCompetitors)
			competitors.GET("/", controllers.GetCompetitors)
			competitors.GET("/categories", controllers.GetCompetitorCategories)
			competitors.GET("/:id", controllers.GetCompetitorByID)
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

				// Withdrawals
				userRoutes.POST("/wallet/withdraw", controllers.CreateWithdrawal)
				userRoutes.POST("/wallet/withdraw/verify", controllers.VerifyWithdrawal)
				userRoutes.GET("/wallet/withdraw/history", controllers.GetWithdrawalHistory)
				userRoutes.GET("/wallet/withdraw/limits", controllers.GetWithdrawalLimits)
				userRoutes.GET("/wallet/withdraw/pending", controllers.GetPendingWithdrawal)
				userRoutes.POST("/wallet/withdraw/cancel", controllers.CancelWithdrawal)

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
				// Rutas duales para evitar redirección
				adminUsers.GET("", controllers.GetUsers)
				adminUsers.GET("/", controllers.GetUsers)
				adminUsers.GET("/:id", controllers.GetUserByID)
				adminUsers.PATCH("/:id/role", controllers.UpdateUserRole)
				adminUsers.PATCH("/:id/status", controllers.UpdateUserStatus)
			}

			// Gestión de Torneos
			adminTournaments := admin.Group("/tournaments")
			{
				// Rutas duales para evitar redirección 307
				adminTournaments.POST("", controllers.CreateTournament)
				adminTournaments.POST("/", controllers.CreateTournament)
				adminTournaments.PATCH("/:id/status", controllers.UpdateTournamentStatus)
			}

			// Gestión de Sesiones
			adminSessions := admin.Group("/sessions")
			{
				adminSessions.POST("", controllers.CreateSession)
				adminSessions.POST("/", controllers.CreateSession)
				adminSessions.PATCH("/:id/status", controllers.UpdateSessionStatus)
			}

			// Gestión de Eventos Globales - Rutas específicas primero
			adminEvents := admin.Group("/events")
			{
				adminEvents.POST("/:event_id/selections", controllers.CreateSelection)
				adminEvents.GET("/:event_id/selections", controllers.GetEventSelections)
				adminEvents.POST("/:event_id/competitors", controllers.SetEventCompetitors)
				adminEvents.POST("/:event_id/settle", controllers.SettleEvent)
				adminEvents.POST("", controllers.CreateGlobalEvent)
				adminEvents.POST("/", controllers.CreateGlobalEvent)
				adminEvents.PUT("/:id", controllers.UpdateEvent)
				adminEvents.DELETE("/:id", controllers.DeleteEvent)
				adminEvents.GET("/available", controllers.GetAvailableEventsForTournament)
			}

			// Gestión de Eventos en Torneos (asignación)
			adminTournamentEvents := admin.Group("/tournament-events")
			{
				adminTournamentEvents.POST("", controllers.AssignEventToTournament)
				adminTournamentEvents.DELETE("/:id", controllers.RemoveEventFromTournament)
			}

			// Gestión de Categorías
			adminCategories := admin.Group("/categories")
			{
				adminCategories.POST("", categoryCtrl.CreateCategory)
				adminCategories.POST("/", categoryCtrl.CreateCategory)
				adminCategories.PUT("", categoryCtrl.UpdateCategory)
				adminCategories.PUT("/:id", categoryCtrl.UpdateCategory)
				adminCategories.DELETE("", categoryCtrl.DeleteCategory)
				adminCategories.DELETE("/:id", categoryCtrl.DeleteCategory)
				adminCategories.PATCH("", categoryCtrl.ToggleCategoryStatus)
				adminCategories.PATCH("/:id/status", categoryCtrl.ToggleCategoryStatus)
			}

			// Gestión de Competidores (catálogo global)
			adminCompetitors := admin.Group("/competitors")
			{
				adminCompetitors.POST("", controllers.CreateCompetitor)
				adminCompetitors.POST("/", controllers.CreateCompetitor)
				adminCompetitors.PUT("/:id", controllers.UpdateCompetitor)
				adminCompetitors.DELETE("/:id", controllers.DeleteCompetitor)
			}
		}
	}

	return r
}

func parseCSVEnv(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

// contains checks if a string is present in a slice of strings.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
