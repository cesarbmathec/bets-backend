package controllers

import (
	"net/http"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

// CreateSelection godoc
// @Summary      Crear una opción de apuesta para un evento
// @Description  Crea una selección (macho, hembra, alta, baja, runline, etc.)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body dtos.CreateSelectionRequest true "Datos de la selección"
// @Success      201 {object} utils.Response{data=models.PickableSelection}
// @Router       /admin/events/selections [post]
// @Security     BearerAuth
func CreateSelection(c *gin.Context) {
	var input dtos.CreateSelectionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar que el evento existe
	var event models.Event
	if err := config.DB.First(&event, input.EventID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "El evento no existe", nil)
		return
	}

	selection := models.PickableSelection{
		EventID:       input.EventID,
		Description:   input.Description,
		SelectionType: input.SelectionType,
		Line:          input.Line,
		CompetitorID:  input.CompetitorID,
		PointsForWin:  input.PointsForWin,
		PointsForPush: input.PointsForPush,
		Status:        "pending",
	}

	if err := config.DB.Create(&selection).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al crear la selección", err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Selección creada correctamente", selection)
}

// GetEventSelections godoc
// @Summary      Listar opciones de apuesta de un evento
// @Tags         events
// @Param        id path int true "ID del Evento"
// @Success      200 {object} utils.Response{data=[]models.PickableSelection}
// @Router       /events/{id}/selections [get]
func GetEventSelections(c *gin.Context) {
	eventID := c.Param("id")
	var selections []models.PickableSelection

	// Buscar selecciones asociadas al evento
	config.DB.Where("event_id = ?", eventID).Find(&selections)

	utils.Success(c, http.StatusOK, "Opciones disponibles", selections)
}
