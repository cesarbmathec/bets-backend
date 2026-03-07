package controllers

import (
	"net/http"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
)

// GetCompetitors godoc
// @Summary      Listar competidores
// @Description  Obtiene todos los competidores del catálogo global
// @Tags         competitors
// @Param        category query string false "Filtrar por categoría"
// @Param        search query string false "Buscar por nombre"
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.Competitor}
// @Router       /competitors [get]
func GetCompetitors(c *gin.Context) {
	var competitors []models.Competitor
	query := config.DB.Model(&models.Competitor{})

	// Filtrar por categoría
	category := c.Query("category")
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Buscar por nombre
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Filtrar solo activos por defecto
	if c.Query("include_inactive") != "true" {
		query = query.Where("status = ?", "active")
	}

	if err := query.Order("category asc, name asc").Find(&competitors).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener competidores", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Competitores obtenidos", competitors)
}

// GetCompetitorByID godoc
// @Summary      Ver competidor por ID
// @Tags         competitors
// @Param        id path int true "ID del Competidor"
// @Produce      json
// @Success      200 {object} utils.Response{data=models.Competitor}
// @Router       /competitors/{id} [get]
func GetCompetitorByID(c *gin.Context) {
	id := c.Param("id")
	var competitor models.Competitor

	if err := config.DB.First(&competitor, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Competidor no encontrado", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Competidor encontrado", competitor)
}

// CreateCompetitor godoc
// @Summary      Crear un competidor
// @Description  Crea un nuevo competidor en el catálogo global
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body dtos.CreateCompetitorRequest true "Datos del competidor"
// @Success      201 {object} utils.Response{data=models.Competitor}
// @Router       /admin/competitors [post]
// @Security     BearerAuth
func CreateCompetitor(c *gin.Context) {
	var input dtos.CreateCompetitorRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar que no existe un competidor con el mismo nombre y categoría
	var existing models.Competitor
	if err := config.DB.Where("name = ? AND category = ?", input.Name, input.Category).First(&existing).Error; err == nil {
		utils.Error(c, http.StatusConflict, "Ya existe un competidor con ese nombre en esta categoría", nil)
		return
	}

	competitor := models.Competitor{
		Name:           input.Name,
		Category:       input.Category,
		AssignedNumber: input.AssignedNumber,
		Description:    input.Description,
		Status:         "active",
	}

	if err := config.DB.Create(&competitor).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al crear competidor", err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Competidor creado con éxito", competitor)
}

// UpdateCompetitor godoc
// @Summary      Actualizar un competidor
// @Tags         admin
// @Param        id path int true "ID del Competidor"
// @Accept       json
// @Produce      json
// @Param        request body dtos.UpdateCompetitorRequest true "Datos a actualizar"
// @Success      200 {object} utils.Response{data=models.Competitor}
// @Router       /admin/competitors/{id} [put]
// @Security     BearerAuth
func UpdateCompetitor(c *gin.Context) {
	id := c.Param("id")
	var competitor models.Competitor

	if err := config.DB.First(&competitor, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Competidor no encontrado", nil)
		return
	}

	var input dtos.UpdateCompetitorRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Actualizar campos
	if input.Name != "" {
		competitor.Name = input.Name
	}
	if input.Category != "" {
		competitor.Category = input.Category
	}
	if input.AssignedNumber > 0 {
		competitor.AssignedNumber = input.AssignedNumber
	}
	if input.Description != "" {
		competitor.Description = input.Description
	}
	if input.Status != "" {
		competitor.Status = input.Status
	}

	if err := config.DB.Save(&competitor).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al actualizar competidor", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Competidor actualizado con éxito", competitor)
}

// DeleteCompetitor godoc
// @Summary      Eliminar un competidor (baja lógica)
// @Description  Cambia el estado del competidor a inactivo
// @Tags         admin
// @Param        id path int true "ID del Competidor"
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /admin/competitors/{id} [delete]
// @Security     BearerAuth
func DeleteCompetitor(c *gin.Context) {
	id := c.Param("id")
	var competitor models.Competitor

	if err := config.DB.First(&competitor, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Competidor no encontrado", nil)
		return
	}

	// Baja lógica - cambiar status a inactive
	competitor.Status = "inactive"
	if err := config.DB.Save(&competitor).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al eliminar competidor", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Competidor eliminado (inactivado)", nil)
}

// GetCompetitorCategories godoc
// @Summary      Listar categorías de competidores
// @Description  Obtiene las categorías únicas de competidores
// @Tags         competitors
// @Produce      json
// @Success      200 {object} utils.Response{data=[]string}
// @Router       /competitors/categories [get]
func GetCompetitorCategories(c *gin.Context) {
	var categories []string
	query := "SELECT DISTINCT category FROM competitors WHERE category IS NOT NULL AND category != '' ORDER BY category asc"

	if err := config.DB.Raw(query).Scan(&categories).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener categorías", err.Error())
		return
	}

	// Si no hay categorías, devolver algunas por defecto
	if len(categories) == 0 {
		categories = []string{"Fútbol", "Caballos", "Béisbol", "Baloncesto", "Fórmula 1", "Tenis", "Boxeo", "Otros"}
	}

	utils.Success(c, http.StatusOK, "Categorías obtenidas", categories)
}

// AddCompetitorToEvent godoc
// @Summary      Agregar un competidor del catálogo a un evento
// @Description  Agrega un competidor existente a un evento específico
// @Tags         admin
// @Param        id path int true "ID del Evento"
// @Accept       json
// @Produce      json
// @Param        request body map[string]interface{} true "competitor_id"
// @Success      200 {object} utils.Response
// @Router       /admin/events/{id}/add-competitor [post]
// @Security     BearerAuth
func AddCompetitorToEvent(c *gin.Context) {
	eventID := c.Param("id")

	// Obtener competitor_id del body
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	competitorID, ok := body["competitor_id"].(float64)
	if !ok {
		utils.Error(c, http.StatusBadRequest, "competitor_id es requerido", nil)
		return
	}

	// Verificar que el evento existe
	var event models.Event
	if err := config.DB.First(&event, eventID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Evento no encontrado", nil)
		return
	}

	// Verificar que el competidor existe
	var competitor models.Competitor
	if err := config.DB.First(&competitor, uint(competitorID)).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Competidor no encontrado", nil)
		return
	}

	// Crear el competidor del evento (copia los datos)
	eventCompetitor := models.EventCompetitor{
		EventID:        event.ID,
		Name:           competitor.Name,
		AssignedNumber: competitor.AssignedNumber,
	}

	if err := config.DB.Create(&eventCompetitor).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al agregar competidor al evento", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Competidor agregado al evento", eventCompetitor)
}
