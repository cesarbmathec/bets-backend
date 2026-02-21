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

// CreateEvent godoc
// @Summary      Crear un evento dentro de un torneo
// @Description  Crea un nuevo evento (partido/carrera) dentro de una sesión
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body dtos.CreateEventRequest true "Datos del evento"
// @Success      201 {object} utils.Response{data=models.Event}
// @Router       /admin/events [post]
// @Security     BearerAuth
func CreateEvent(c *gin.Context) {
	var input dtos.CreateEventRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	event := models.Event{
		TournamentID: input.TournamentID,
		Name:         input.Name,
		Order:        input.Order,
		StartTime:    input.StartTime,
		Status:       "scheduled",
	}

	if err := config.DB.Create(&event).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al crear evento", err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Evento creado con éxito", event)
}

// GetEventBySlug godoc
// @Summary      Ver detalle de evento por Slug (incluye competidores)
// @Tags         events
// @Param        slug path string true "Slug del Evento"
// @Router       /events/s/{slug} [get]
// @Success      200 {object} utils.Response{data=models.Event}
// @Security     BearerAuth
func GetEventBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var event models.Event

	// Preload carga automáticamente los competidores asociados
	if err := config.DB.Preload("Competitors").Where("slug = ?", slug).First(&event).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Evento no encontrado", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Evento encontrado", event)
}

// GetEventByID godoc
// @Summary      Ver detalle de evento por ID
// @Tags         events
// @Param        id path int true "ID del Evento"
// @Produce      json
// @Success      200 {object} utils.Response{data=models.Event}
// @Router       /events/id/{id} [get]
func GetEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	// Preload Competitors para ver quiénes participan en el evento
	if err := config.DB.Preload("Competitors").First(&event, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Evento no encontrado", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Evento encontrado", event)
}

// SetEventCompetitors godoc
// @Summary      Asignar competidores/caballos a un evento
// @Description  Asocia equipos o caballos a un evento
// @Tags         admin
// @Param        id path int true "ID del Evento"
// @Param        request body dtos.SetCompetitorsRequest true "Lista de competidores"
// @Router       /admin/events/{id}/competitors [post]
// @Security     BearerAuth
func SetEventCompetitors(c *gin.Context) {
	eventID := c.Param("id")
	var input dtos.SetCompetitorsRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos de competidores inválidos", err.Error())
		return
	}

	tx := config.DB.Begin()

	// Opcional: Limpiar competidores previos si es una re-asignación
	tx.Where("event_id = ?", eventID).Delete(&models.EventCompetitor{})

	for _, comp := range input.Competitors {
		newComp := models.EventCompetitor{
			EventID:        utils.StringToUint(eventID),
			Name:           comp.Name,
			AssignedNumber: comp.AssignedNumber,
		}
		if err := tx.Create(&newComp).Error; err != nil {
			tx.Rollback()
			utils.Error(c, http.StatusInternalServerError, "Error al guardar competidores", err.Error())
			return
		}
	}

	tx.Commit()
	utils.Success(c, http.StatusOK, "Competidores asignados correctamente", nil)
}
