// @title			Sistema de Apuestas API
// @version			1.0
// @description		API para la gestión de torneos y apuestas, usuarios y transacciones.
// @termsOfService	http://swagger.io/terms/

// @contact.name	Soporte Técnico
// @contact.email	soporte@tuapp.com

// @host			localhost:8080
// @BasePath		/api/v1

// @securityDefinitions.apikey BearerAuth
// @in				header
// @name			Authorization
// @description Escribe 'Bearer ' seguido de tu token JWT

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/migrations"
	"github.com/cesarbmathec/bets-backend/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Archivo .env no encontrado, usando variables de entorno del sistema")
	}

	// Configurar modo de Gin
	gin.SetMode(os.Getenv("GIN_MODE"))

	// Inicializar Base de Datos
	config.ConnectDatabase()
	db := config.GetDB()

	// Ejecutar Migraciones y Seeds iniciales
	migrations.RunMigrations(db)

	// Configurar el Router
	r := routes.SetupRouter()

	//Servidores Proxy de confianza (Solo para producción)
	if os.Getenv("GIN_MODE") == "release" {
		trustedProxies := strings.Split(os.Getenv("TRUSTED_PROXIES"), ",")
		r.SetTrustedProxies(trustedProxies)
	}

	// Iniciar el servidor
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("\n🚀 Servidor de Apuestas corriendo en: http://localhost:%s\n", port)
	fmt.Printf("📄 Documentación Swagger: http://localhost:%s/swagger/index.html\n\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatal("❌ Error al iniciar el servidor: ", err)
	}
}
