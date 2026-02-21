package controllers

import (
	"net/http"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
)

// GetUsers godoc
// @Summary      Listar usuarios
// @Description  Obtiene todos los usuarios del sistema (solo admin)
// @Tags         admin
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=[]models.User}
// @Router       /admin/users [get]
func GetUsers(c *gin.Context) {
	var users []models.User

	if err := config.DB.Find(&users).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener usuarios", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Lista de usuarios", users)
}

// GetUserByID godoc
// @Summary      Ver usuario por ID
// @Description  Obtiene los detalles de un usuario específico
// @Tags         admin
// @Security     BearerAuth
// @Param        id path int true "ID del Usuario"
// @Success      200 {object} utils.Response{data=models.User}
// @Router       /admin/users/{id} [get]
func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := config.DB.First(&user, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Usuario no encontrado", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Usuario encontrado", user)
}

// UpdateUserRole godoc
// @Summary      Actualizar rol de usuario
// @Description  Permite al admin cambiar el rol de un usuario (user o admin)
// @Tags         admin
// @Security     BearerAuth
// @Param        id path int true "ID del Usuario"
// @Param        request body dtos.UpdateUserRoleRequest true "Nuevo rol"
// @Success      200 {object} utils.Response
// @Router       /admin/users/{id}/role [patch]
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var input dtos.UpdateUserRoleRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Usuario no encontrado", nil)
		return
	}

	// No permitir que un admin se degrade a sí mismo
	currentUserID, _ := c.Get("userID")
	if currentUserID.(uint) == user.ID && input.Role != "admin" {
		utils.Error(c, http.StatusForbidden, "No puede degradar su propio rol de administrador", nil)
		return
	}

	user.Role = input.Role
	if err := config.DB.Save(&user).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al actualizar rol", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Rol actualizado correctamente", user)
}

// UpdateUserStatus godoc
// @Summary      Actualizar estado de usuario
// @Description  Activa o desactiva un usuario
// @Tags         admin
// @Security     BearerAuth
// @Param        id path int true "ID del Usuario"
// @Param        request body dtos.UpdateUserStatusRequest true "Nuevo estado"
// @Success      200 {object} utils.Response
// @Router       /admin/users/{id}/status [patch]
func UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	var input dtos.UpdateUserStatusRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Usuario no encontrado", nil)
		return
	}

	// No permitir que un admin se desactive a sí mismo
	currentUserID, _ := c.Get("userID")
	if currentUserID.(uint) == user.ID && !input.IsActive {
		utils.Error(c, http.StatusForbidden, "No puede desactivarse a sí mismo", nil)
		return
	}

	user.IsActive = input.IsActive
	if err := config.DB.Save(&user).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al actualizar estado", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Estado actualizado correctamente", user)
}

// GetMyProfile godoc
// @Summary      Ver mi perfil
// @Description  Obtiene el perfil del usuario autenticado
// @Tags         users
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=models.User}
// @Router       /me [get]
func GetMyProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	var user models.User

	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Usuario no encontrado", nil)
		return
	}

	// Ocultar contraseña
	user.Password = ""

	utils.Success(c, http.StatusOK, "Perfil del usuario", user)
}
