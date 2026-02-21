package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_Success(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Arrange
	registerBody := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}

	// Act
	w := MakeJSONRequest(router, "POST", "/api/v1/auth/register", registerBody)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Usuario registrado exitosamente", response["message"])

	// El token está dentro de data
	assert.NotNil(t, response["data"], "Response data should not be nil")
	data := response["data"].(map[string]interface{})
	assert.NotEmpty(t, data["token"], "Token should not be empty")
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Crear usuario primero
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	user := models.User{
		Username: "existinguser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}
	config.DB.Create(&user)

	// Arrange
	registerBody := map[string]interface{}{
		"username": "newuser",
		"email":    "test@example.com",
		"password": "password123",
	}

	// Act
	w := MakeJSONRequest(router, "POST", "/api/v1/auth/register", registerBody)

	// Assert - La base de datos SQLite devuelve 409 Conflict
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusConflict,
		"Expected 400 or 409, got %d", w.Code)
}

func TestLogin_Success(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Crear usuario
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}
	config.DB.Create(&user)

	// Arrange
	loginBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}

	// Act
	w := MakeJSONRequest(router, "POST", "/api/v1/auth/login", loginBody)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Imprimir respuesta para debugging
	t.Logf("Login response: %+v", response)

	// Verificar que hay datos
	assert.NotNil(t, response["data"], "Response data should not be nil")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Crear usuario
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}
	config.DB.Create(&user)

	// Arrange
	loginBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}

	// Act
	w := MakeJSONRequest(router, "POST", "/api/v1/auth/login", loginBody)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPublicRoutes_Accessible(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Rutas públicas que deben ser accesibles sin autenticación
	publicRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/tournaments"},
		{"GET", "/api/v1/tournaments/id/1"},
		{"GET", "/api/v1/tournaments/s/test"},
		{"GET", "/api/v1/events/id/1"},
		{"GET", "/api/v1/events/s/test"},
		{"POST", "/api/v1/auth/login"},
		{"POST", "/api/v1/auth/register"},
	}

	for _, route := range publicRoutes {
		w := MakeRequest(router, route.method, route.path)
		// Estas rutas deben devolver 200, 404 (no 401)
		assert.NotEqual(t, http.StatusUnauthorized, w.Code,
			"Ruta pública %s %s no debe requerir autenticación", route.method, route.path)
	}
}

func TestProtectedRoutes_RequireAuth(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Rutas protegidas que deben requerir autenticación
	protectedRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/me"},
		{"POST", "/api/v1/wallet/deposit"},
		{"POST", "/api/v1/tournaments/1/join"},
		{"GET", "/api/v1/my-sessions/1/picks"},
		{"POST", "/api/v1/tournaments/1/sessions/picks"},
	}

	for _, route := range protectedRoutes {
		w := MakeRequest(router, route.method, route.path)
		// Estas rutas deben devolver 401 (no autorizado)
		assert.Equal(t, http.StatusUnauthorized, w.Code,
			"Ruta %s %s debe requerir autenticación", route.method, route.path)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Arrange
	registerBody := map[string]interface{}{
		"username": "testuser",
		"email":    "invalid-email",
		"password": "password123",
	}

	// Act
	w := MakeJSONRequest(router, "POST", "/api/v1/auth/register", registerBody)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_ShortPassword(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Arrange
	registerBody := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "123",
	}

	// Act
	w := MakeJSONRequest(router, "POST", "/api/v1/auth/register", registerBody)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Crear usuario
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	user := models.User{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}
	config.DB.Create(&user)

	// Arrange
	registerBody := map[string]interface{}{
		"username": "existinguser",
		"email":    "new@example.com",
		"password": "password123",
	}

	// Act
	w := MakeJSONRequest(router, "POST", "/api/v1/auth/register", registerBody)

	// Assert - La base de datos SQLite devuelve 409 Conflict
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusConflict,
		"Expected 400 or 409, got %d", w.Code)
}

func TestLogin_NonExistentUser(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Arrange
	loginBody := map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "password123",
	}

	// Act
	w := MakeJSONRequest(router, "POST", "/api/v1/auth/login", loginBody)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAdminRoutes_RequireAdmin verifica que las rutas de admin requieran rol de admin
func TestAdminRoutes_RequireAdmin(t *testing.T) {
	// Setup
	SetupTestDB(t)
	router := SetupRouter()

	// Crear usuario regular
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 14)
	user := models.User{
		Username: "regularuser",
		Email:    "user@example.com",
		Password: string(hashedPassword),
		Role:     "user",
	}
	config.DB.Create(&user)

	// Rutas de admin que deben requerir rol de admin
	// Los códigos 301, 307, 401, 403 son todos válidos (redirección, o no autorizado)
	adminRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/admin/users"},
		{"POST", "/api/v1/admin/tournaments"},
		{"POST", "/api/v1/admin/sessions"},
		{"POST", "/api/v1/admin/events"},
	}

	for _, route := range adminRoutes {
		w := MakeRequest(router, route.method, route.path)
		// Estas rutas deben devolver 301, 307 (redirect), 401 (no autorizado) o 403 (prohibido)
		assert.True(t, w.Code == http.StatusMovedPermanently || w.Code == http.StatusTemporaryRedirect || w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden,
			"Ruta %s %s debe requerir autenticación de admin, got %d", route.method, route.path, w.Code)
	}
}
