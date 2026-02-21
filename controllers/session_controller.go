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

// CreateSession godoc
// @Summary      Crear una sesión dentro de un torneo
// @Description  Crea una nueva sesión (jornada/día) dentro de un torneo existente
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body dtos.CreateSessionRequest true "Datos de la sesión"
// @Success      201 {object} utils.Response{data=models.Session}
// @Router       /admin/sessions [post]
// @Security     BearerAuth
func CreateSession(c *gin.Context) {
	var input dtos.CreateSessionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar que el torneo existe
	var tournament models.Tournament
	if err := config.DB.First(&tournament, input.TournamentID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Torneo no encontrado", nil)
		return
	}

	// Validar que no existe ya una sesión con ese número para el torneo
	var existingSession models.Session
	if err := config.DB.Where("tournament_id = ? AND session_number = ?", input.TournamentID, input.SessionNumber).First(&existingSession).Error; err == nil {
		utils.Error(c, http.StatusConflict, fmt.Sprintf("Ya existe la sesión #%d para este torneo", input.SessionNumber), nil)
		return
	}

	session := models.Session{
		TournamentID:  input.TournamentID,
		SessionNumber: input.SessionNumber,
		StartTime:     input.StartTime,
		EndTime:       input.EndTime,
		Description:   input.Description,
		Status:        "open",
	}

	if err := config.DB.Create(&session).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al crear la sesión", err.Error())
		return
	}

	utils.Success(c, http.StatusCreated, "Sesión creada correctamente", session)
}

// GetTournamentSessions godoc
// @Summary      Listar sesiones de un torneo
// @Description  Obtiene todas las sesiones de un torneo con información de eventos y predicciones
// @Tags         sessions
// @Param        id path int true "ID del Torneo"
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.Session}
// @Router       /tournaments/{id}/sessions [get]
func GetTournamentSessions(c *gin.Context) {
	tournamentID := c.Param("id")
	var sessions []models.Session

	if err := config.DB.Preload("Events").
		Where("tournament_id = ?", tournamentID).
		Order("session_number asc").
		Find(&sessions).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener sesiones", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Sesiones del torneo", sessions)
}

// GetSessionByID godoc
// @Summary      Ver detalle de una sesión
// @Description  Obtiene los detalles de una sesión específica
// @Tags         sessions
// @Param        id path int true "ID de la Sesión"
// @Produce      json
// @Success      200 {object} utils.Response{data=models.Session}
// @Router       /sessions/{id} [get]
func GetSessionByID(c *gin.Context) {
	sessionID := c.Param("id")
	var session models.Session

	if err := config.DB.Preload("Events.PickableSelections").
		Preload("Events.Competitors").
		First(&session, sessionID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Sesión no encontrada", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Sesión encontrada", session)
}

// UpdateSessionStatus godoc
// @Summary      Actualizar estado de una sesión
// @Description  Cambia el estado de una sesión (open, closed, settled)
// @Tags         admin
// @Param        id path int true "ID de la Sesión"
// @Param        request body dtos.UpdateSessionStatusRequest true "Nuevo estado"
// @Success      200 {object} utils.Response
// @Router       /admin/sessions/{id}/status [patch]
// @Security     BearerAuth
func UpdateSessionStatus(c *gin.Context) {
	sessionID := c.Param("id")
	var input dtos.UpdateSessionStatusRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	var session models.Session
	if err := config.DB.First(&session, sessionID).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Sesión no encontrada", nil)
		return
	}

	session.Status = input.Status
	if err := config.DB.Save(&session).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al actualizar estado", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Estado de sesión actualizado", session)
}

// SubmitPicks godoc
// @Summary      Enviar predicciones para una sesión
// @Description  Permite a un participante enviar sus selecciones para una sesión específica
// @Tags         users
// @Security     ApiKeyAuth
// @Param        id path int true "ID del Torneo"
// @Param        request body dtos.SubmitPicksBySessionRequest true "Datos de las predicciones"
// @Success      201 {object} utils.Response
// @Router       /tournaments/{id}/sessions/picks [post]
func SubmitPicksBySession(c *gin.Context) {
	tournamentID := c.Param("id")
	userID, _ := c.Get("userID")

	var input dtos.SubmitPicksBySessionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	tx := config.DB.Begin()

	// 1. Validar que el usuario es participante del torneo
	var participant models.TournamentParticipant
	if err := tx.Where("user_id = ? AND tournament_id = ?", userID, tournamentID).First(&participant).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusForbidden, "No estás inscrito en este torneo", nil)
		return
	}

	// 2. Validar que la sesión existe y está abierta
	var session models.Session
	if err := tx.First(&session, input.SessionID).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusNotFound, "Sesión no encontrada", nil)
		return
	}

	if session.TournamentID != utils.StringToUint(tournamentID) {
		tx.Rollback()
		utils.Error(c, http.StatusBadRequest, "La sesión no pertenece a este torneo", nil)
		return
	}

	if session.Status != "open" {
		tx.Rollback()
		utils.Error(c, http.StatusBadRequest, "La sesión no está abierta para predicciones", nil)
		return
	}

	// 3. Validar que no haya pasado la hora límite de la sesión
	if time.Now().After(session.EndTime) {
		tx.Rollback()
		utils.Error(c, http.StatusBadRequest, "Ya cerró la hora límite para hacer selecciones en esta sesión", nil)
		return
	}

	// 4. Obtener configuración del torneo
	var tournament models.Tournament
	if err := tx.First(&tournament, tournamentID).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusNotFound, "Torneo no encontrado", nil)
		return
	}

	// 5. Validar cantidad de selecciones requeridas
	requiredSelections := tournament.Settings.SelectionsPerSession
	if len(input.SelectionIDs) != requiredSelections {
		tx.Rollback()
		utils.Error(c, http.StatusBadRequest,
			fmt.Sprintf("Debe hacer exactamente %d selecciones por sesión", requiredSelections), nil)
		return
	}

	// 6. Validar tipos de selección obligatorios (si están configurados)
	if len(tournament.Settings.RequiredSelectionTypes) > 0 {
		for i, selectionID := range input.SelectionIDs {
			var sel models.PickableSelection
			if err := tx.First(&sel, selectionID).Error; err != nil {
				tx.Rollback()
				utils.Error(c, http.StatusNotFound, fmt.Sprintf("Selección #%d no encontrada", selectionID), nil)
				return
			}

			if i < len(tournament.Settings.RequiredSelectionTypes) {
				requiredType := tournament.Settings.RequiredSelectionTypes[i]
				if sel.SelectionType != requiredType {
					tx.Rollback()
					utils.Error(c, http.StatusBadRequest,
						fmt.Sprintf("La selección #%d debe ser de tipo '%s' (posición %d)", i+1, requiredType, i+1), nil)
					return
				}
			}
		}
	}

	var savedPicks []models.UserPick

	// 7. Procesar cada selección
	for _, selectionID := range input.SelectionIDs {
		var selection models.PickableSelection
		// Preload Event para validar fechas
		if err := tx.Preload("Event").First(&selection, selectionID).Error; err != nil {
			continue
		}

		// Validar que el evento de la selección pertenezca al torneo actual
		tournamentIDUint := utils.StringToUint(tournamentID)
		if selection.Event.TournamentID != tournamentIDUint {
			tx.Rollback()
			utils.Error(c, http.StatusBadRequest, "La selección no pertenece a este torneo", nil)
			return
		}

		// Validar que el evento no haya comenzado
		if time.Now().After(selection.Event.StartTime) {
			tx.Rollback()
			utils.Error(c, http.StatusBadRequest, "El evento ya ha comenzado", nil)
			return
		}

		// Buscar si ya existe un pick para esta selección por este participante en esta sesión
		var existingPick models.UserPick
		err := tx.Where("participant_id = ? AND selection_id = ? AND session_id = ?",
			participant.ID, selection.ID, session.ID).First(&existingPick).Error

		if err == nil {
			// Ya existe, actualizar
			existingPick.SessionID = session.ID
			tx.Save(&existingPick)
			savedPicks = append(savedPicks, existingPick)
		} else {
			// Crear nuevo pick
			newPick := models.UserPick{
				ParticipantID: participant.ID,
				SelectionID:   selection.ID,
				SessionID:     session.ID,
				Status:        "pending",
			}
			if err := tx.Create(&newPick).Error; err != nil {
				tx.Rollback()
				utils.Error(c, http.StatusInternalServerError, "Error al guardar predicción", err.Error())
				return
			}
			savedPicks = append(savedPicks, newPick)
		}
	}

	tx.Commit()
	utils.Success(c, http.StatusCreated, "Predicciones guardadas para la sesión", savedPicks)
}

// GetSessionPicks godoc
// @Summary      Ver predicciones de un usuario en una sesión
// @Description  Obtiene las predicciones realizadas por el usuario en una sesión específica
// @Tags         users
// @Security     ApiKeyAuth
// @Param        session_id path int true "ID de la Sesión"
// @Success      200 {object} utils.Response{data=[]models.UserPick}
// @Router       /my-sessions/{session_id}/picks [get]
func GetSessionPicks(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID, _ := c.Get("userID")

	// Buscar el participant
	var participant models.TournamentParticipant
	if err := config.DB.Where("user_id = ?", userID).First(&participant).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "No estás inscrito en ningún torneo", nil)
		return
	}

	var picks []models.UserPick
	if err := config.DB.Preload("Selection").
		Where("participant_id = ? AND session_id = ?", participant.ID, sessionID).
		Find(&picks).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener predicciones", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Tus predicciones en esta sesión", picks)
}
