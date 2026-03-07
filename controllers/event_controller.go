package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

// ==================== ENDPOINTS PÚBLICOS ====================

// GetGlobalEvents godoc
// @Summary      Listar eventos globales
// @Description  Obtiene todos los eventos disponibles (no asignados a un torneo específico)
// @Tags         events
// @Param        search query string false "Buscar por nombre"
// @Param        status query string false "Filtrar por estado"
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.Event}
// @Router       /events [get]
func GetGlobalEvents(c *gin.Context) {
	var events []models.Event
	query := config.DB // Todos los eventos son globales ahora (sin TournamentID directo)

	// Buscar por nombre
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	// Filtrar por estado
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("start_time asc").Preload("Competitors").Find(&events).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener eventos", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Eventos obtenidos", events)
}

// GetEventByID godoc
// @Summary      Ver detalle de evento por ID
// @Tags         events
// @Param        id path int true "ID del Evento"
// @Produce      json
// @Success      200 {object} utils.Response{data=models.Event}
// @Router       /events/{id} [get]
func GetEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := config.DB.Preload("Competitors").First(&event, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Evento no encontrado", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Evento encontrado", event)
}

// GetTournamentEventsByTournament godoc
// @Summary      Listar eventos de un torneo
// @Description  Obtiene los eventos asignados a un torneo específico
// @Tags         events
// @Param        tournament_id path int true "ID del Torneo"
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.TournamentEvent}
// @Router       /events/tournament/{tournament_id} [get]
func GetTournamentEventsByTournament(c *gin.Context) {
	tournamentID := c.Param("tournament_id")

	var tournamentEvents []models.TournamentEvent
	if err := config.DB.
		Preload("Event.Competitors").
		Preload("Session").
		Where("tournament_id = ?", tournamentID).
		Order("session_id asc, \"order\" asc").
		Find(&tournamentEvents).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener eventos del torneo", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Eventos del torneo", tournamentEvents)
}

// ==================== ENDPOINTS DE ADMIN ====================

// CreateGlobalEvent godoc
// @Summary      Crear evento global
// @Description  Crea un evento sin asignarlo a un torneo (puede asignarse después)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body dtos.CreateEventRequest true "Datos del evento"
// @Success      201 {object} utils.Response{data=models.Event}
// @Router       /admin/events [post]
// @Security     BearerAuth
func CreateGlobalEvent(c *gin.Context) {
	var input dtos.CreateEventRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	fmt.Printf("Creating event: %+v\n", input)

	// Parsear fecha
	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Formato de fecha inválido", err.Error())
		return
	}

	event := models.Event{
		Name:      input.Name,
		Venue:     input.Venue,
		Line:      input.Line,
		StartTime: startTime,
		Status:    "scheduled",
		// TournamentID queda NULL - es un evento global
	}

	if err := config.DB.Create(&event).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al crear evento", err.Error())
		return
	}

	fmt.Printf("Event created with ID: %d\n", event.ID)

	config.DB.Preload("Competitors").First(&event, event.ID)
	utils.Success(c, http.StatusCreated, "Evento creado con éxito", event)
}

// UpdateEvent godoc
// @Summary      Actualizar evento
// @Description  Actualiza un evento existente
// @Tags         admin
// @Param        id path int true "ID del Evento"
// @Accept       json
// @Produce      json
// @Param        request body dtos.UpdateEventRequest true "Datos a actualizar"
// @Success      200 {object} utils.Response{data=models.Event}
// @Router       /admin/events/{id} [put]
// @Security     BearerAuth
func UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := config.DB.First(&event, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Evento no encontrado", nil)
		return
	}

	var input dtos.UpdateEventRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Actualizar campos
	if input.Name != "" {
		event.Name = input.Name
	}
	if input.Venue != "" {
		event.Venue = input.Venue
	}
	if input.Line > 0 {
		event.Line = input.Line
	}
	if input.Status != "" {
		event.Status = input.Status
	}
	if input.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, input.StartTime)
		if err == nil {
			event.StartTime = startTime
		}
	}

	if err := config.DB.Save(&event).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al actualizar evento", err.Error())
		return
	}

	config.DB.Preload("Competitors").First(&event, event.ID)
	utils.Success(c, http.StatusOK, "Evento actualizado", event)
}

// DeleteEvent godoc
// @Summary      Eliminar evento
// @Description  Elimina un evento (solo si no tiene picks asociados)
// @Tags         admin
// @Param        id path int true "ID del Evento"
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /admin/events/{id} [delete]
// @Security     BearerAuth
func DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := config.DB.First(&event, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Evento no encontrado", nil)
		return
	}

	// Verificar si tiene selecciones con picks
	var pickCount int64
	config.DB.Model(&models.PickableSelection{}).Where("event_id = ?", id).Count(&pickCount)
	if pickCount > 0 {
		utils.Error(c, http.StatusConflict, "No se puede eliminar el evento porque tiene selecciones activas", nil)
		return
	}

	// Eliminar competidores del evento
	config.DB.Where("event_id = ?", id).Delete(&models.EventCompetitor{})
	// Eliminar selecciones
	config.DB.Where("event_id = ?", id).Delete(&models.PickableSelection{})
	// Eliminar de tournament_events
	config.DB.Where("event_id = ?", id).Delete(&models.TournamentEvent{})
	// Eliminar evento
	if err := config.DB.Delete(&event).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al eliminar evento", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Evento eliminado", nil)
}

// AssignEventToTournament godoc
// @Summary      Asignar evento a un torneo
// @Description  Asigna un evento existente a un torneo y sesión
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body dtos.AssignEventToTournamentRequest true "Datos de asignación"
// @Success      201 {object} utils.Response{data=models.TournamentEvent}
// @Router       /admin/tournament-events [post]
// @Security     BearerAuth
func AssignEventToTournament(c *gin.Context) {
	var input dtos.AssignEventToTournamentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Verificar que el evento existe
	var event models.Event
	if err := config.DB.First(&event, input.EventID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Evento no encontrado", nil)
		return
	}

	// Verificar que el torneo existe
	var tournament models.Tournament
	if err := config.DB.First(&tournament, input.TournamentID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Torneo no encontrado", nil)
		return
	}

	// Si se específica sesión, verificar que existe y pertenece al torneo
	if input.SessionID != nil && *input.SessionID > 0 {
		var session models.Session
		if err := config.DB.First(&session, *input.SessionID).Error; err != nil {
			utils.Error(c, http.StatusNotFound, "Sesión no encontrada", nil)
			return
		}
		if session.TournamentID != input.TournamentID {
			utils.Error(c, http.StatusBadRequest, "La sesión no pertenece a este torneo", nil)
			return
		}
	}

	// Verificar si ya está asignado
	var existing models.TournamentEvent
	if err := config.DB.Where("tournament_id = ? AND event_id = ?", input.TournamentID, input.EventID).First(&existing).Error; err == nil {
		// Ya existe, actualizar
		existing.SessionID = input.SessionID
		existing.Order = input.Order
		config.DB.Save(&existing)
		utils.Success(c, http.StatusOK, "Asignación actualizada", existing)
		return
	}

	// Crear nueva asignación
	tournamentEvent := models.TournamentEvent{
		TournamentID: input.TournamentID,
		EventID:      input.EventID,
		SessionID:    input.SessionID,
		Order:        input.Order,
	}

	if err := config.DB.Create(&tournamentEvent).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al asignar evento", err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Evento asignado al torneo", tournamentEvent)
}

// RemoveEventFromTournament godoc
// @Summary      Quitar evento de un torneo
// @Description  Elimina la asignación de un evento a un torneo
// @Tags         admin
// @Param        id path int true "ID de la relación TournamentEvent"
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /admin/tournament-events/{id} [delete]
// @Security     BearerAuth
func RemoveEventFromTournament(c *gin.Context) {
	id := c.Param("id")
	var tournamentEvent models.TournamentEvent

	if err := config.DB.First(&tournamentEvent, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Asignación no encontrada", nil)
		return
	}

	if err := config.DB.Delete(&tournamentEvent).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al quitar evento del torneo", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Evento removido del torneo", nil)
}

// SetEventCompetitors godoc
// @Summary      Establecer competidores de un evento
// @Description  Asigna los competidores (equipos/caballos) a un evento
// @Tags         admin
// @Param        id path int true "ID del Evento"
// @Param        request body dtos.SetEventCompetitorsRequest true "Lista de competidores"
// @Router       /admin/events/{id}/competitors [post]
// @Security     BearerAuth
func SetEventCompetitors(c *gin.Context) {
	eventID := c.Param("event_id")
	var input dtos.SetEventCompetitorsRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos de competidores inválidos", err.Error())
		return
	}

	tx := config.DB.Begin()

	// Eliminar competidores previos
	tx.Where("event_id = ?", eventID).Delete(&models.EventCompetitor{})

	for _, comp := range input.Competitors {
		newComp := models.EventCompetitor{
			EventID:        utils.StringToUint(eventID),
			CompetitorID:   comp.CompetitorID,
			Name:           comp.Name,
			AssignedNumber: comp.AssignedNumber,
			Odds:           comp.Odds,
			Runline:        float64(comp.Runline),
			SuperRunline:   float64(comp.SuperRunline),
			IsFavorite:     comp.IsFavorite,
		}
		if err := tx.Create(&newComp).Error; err != nil {
			tx.Rollback()
			utils.Error(c, http.StatusInternalServerError, "Error al guardar competidores", err.Error())
			return
		}
	}

	tx.Commit()

	// Cargar y retornar competidores actualizados
	var competitors []models.EventCompetitor
	config.DB.Where("event_id = ?", eventID).Find(&competitors)

	utils.Success(c, http.StatusOK, "Competidores actualizados", competitors)
}

// GetAvailableEventsForTournament godoc
// @Summary      Listar eventos disponibles para asignar
// @Description  Obtiene eventos que aún no están asignados a un torneo o están disponibles
// @Tags         admin
// @Param        tournament_id query int true "ID del Torneo"
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.Event}
// @Router       /admin/events/available [get]
// @Security     BearerAuth
func GetAvailableEventsForTournament(c *gin.Context) {
	tournamentID := c.Query("tournament_id")

	fmt.Printf("GetAvailableEventsForTournament called with tournament_id: %s\n", tournamentID)

	// Obtener IDs de eventos ya asignados a este torneo
	var assignedEventIDs []uint
	config.DB.Model(&models.TournamentEvent{}).
		Where("tournament_id = ?", tournamentID).
		Pluck("event_id", &assignedEventIDs)

	fmt.Printf("Assigned event IDs: %v\n", assignedEventIDs)

	// Obtener todos los eventos globales (ya no tienen tournament_id directo)
	// Mostrar TODOS los eventos, no solo los no asignados
	var events []models.Event
	query := config.DB.Where("status != ?", "cancelled")

	if err := query.Order("start_time asc").Preload("Competitors").Find(&events).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener eventos", err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Eventos disponibles", events)
}
