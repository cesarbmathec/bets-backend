package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/routes"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	// Configurar JWT_SECRET para tests
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only-12345")
}

// SetupTestDB crea una base de datos SQLite en memoria para testing
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate
	db.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Transaction{},
		&models.Payment{},
		&models.Event{},
		&models.EventCompetitor{},
		&models.Tournament{},
		&models.TournamentParticipant{},
		&models.UserPick{},
		&models.PickableSelection{},
		&models.Session{},
	)

	// Reemplazar la base de datos global
	config.DB = db

	return db
}

// SetupRouter crea el router para testing
func SetupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return routes.SetupRouter()
}

// MakeRequest hace un request HTTP sin body
func MakeRequest(router *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// MakeJSONRequest hace un request HTTP con body JSON
func MakeJSONRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// MakeAuthRequest hace un request HTTP con autenticación
func MakeAuthRequest(router *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
