package controllers

import (
	"net/http"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

// Login godoc
// @Summary     Iniciar sesión
// @Description Autentica al usuario y devuelve un token JWT
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body dtos.LoginRequest true "Credenciales"
// @Success     200 {object} utils.Response{data=dtos.LoginResponse} "Login exitoso"
// @Failure		400 {object} utils.Response "Datos de entrada inválidos"
// @Failure		401 {object} utils.Response "Credenciales incorrectas"
// @Failure		403 {object} utils.Response "Cuenta desactivada"
// @Router      /auth/login [post]
// @example request -json {"email": "admin@betsystem.com", "password": "Admin123!"}
// @example response -json {"success": true, "message": "Bienvenido al sistema", "data": {"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", "user": {"id": 1, "username": "admin", "email": "admin@betsystem.com", "role": "admin"}}}
func Login(c *gin.Context) {
	var input dtos.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Validar que proporcione email o username
	if input.Email == "" && input.Username == "" {
		utils.Error(c, http.StatusBadRequest, "Debe proporcionar email o username", nil)
		return
	}

	var user models.User
	db := config.GetDB()

	// Buscar por email o username
	query := db.Preload("Wallet")
	if input.Email != "" {
		query = query.Where("email = ?", input.Email)
	} else {
		query = query.Where("username = ?", input.Username)
	}

	if err := query.First(&user).Error; err != nil {
		utils.Error(c, http.StatusUnauthorized, "Credenciales incorrectas", nil)
		return
	}

	// Verificar password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, "Credenciales incorrectas", nil)
		return
	}

	if !user.IsActive {
		utils.Error(c, http.StatusForbidden, "Cuenta de usuario desactivada", nil)
		return
	}

	// Generar Token JWT
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error generando acceso", nil)
		return
	}

	response := dtos.LoginResponse{
		Token: token,
		User: dtos.UserSummary{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}

	utils.Success(c, http.StatusOK, "Bienvenido al sistema", response)
}

// Register godoc
// @Summary      Registro de usuario
// @Description  Crea un nuevo apostador y le asigna una billetera vacía
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dtos.RegisterRequest true "Datos de registro"
// @Success      201 {object} utils.Response "Usuario creado exitosamente"
// @Failure      400 {object} utils.Response "Error de validación"
// @Failure      409 {object} utils.Response "Usuario o email ya existe"
// @Router       /auth/register [post]
// @example request -json {"username": "jugador1", "email": "jugador1@example.com", "password": "Pass123!"}
// @example response -json {"success": true, "message": "Usuario registrado exitosamente", "data": {"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", "user": "jugador1"}}
func Register(c *gin.Context) {
	var input dtos.RegisterRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Error de validación", err.Error())
		return
	}

	db := config.GetDB()

	// Iniciar transacción de base de datos para asegurar integridad (User + Wallet)
	tx := db.Begin()

	user := models.User{
		Username:   input.Username,
		Email:      input.Email,
		Role:       "user",
		IsActive:   true,
		FullName:   input.FullName,
		Phone:      input.Phone,
		DocumentID: input.DocumentID,
	}
	user.HashPassword(input.Password)

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusConflict, "El usuario o email ya existe", nil)
		return
	}

	// Crear billetera inicial para el usuario
	wallet := models.Wallet{
		UserID:       user.ID,
		Balance:      0,
		BonusBalance: 0,
		TokenBalance: 0,
	}

	if err := tx.Create(&wallet).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusInternalServerError, "No se pudo inicializar la billetera", nil)
		return
	}

	tx.Commit()

	token, _ := utils.GenerateToken(user.ID, user.Username, user.Role)

	utils.Success(c, http.StatusCreated, "Usuario registrado exitosamente", gin.H{
		"token": token,
		"user":  user.Username,
	})
}
