package controllers

import (
	"net/http"
	"time"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

// SubmitPicks godoc
// @Summary      Enviar predicciones para un torneo (versión anterior)
// @Description  Permite a un participante enviar una o varias selecciones
// @Tags         users
// @Security     BearerAuth
// @Param        id path int true "ID del Torneo"
// @Param        request body dtos.SubmitPicksRequest true "IDs de las selecciones"
// @Success      201 {object} utils.Response
// @Router       /tournaments/{id}/picks [post]
func SubmitPicks(c *gin.Context) {
	tournamentID := c.Param("id")
	userID, _ := c.Get("userID")

	var input dtos.SubmitPicksRequest
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

	var savedPicks []models.UserPick

	// 2. Procesar cada selección
	for _, selectionID := range input.SelectionIDs {
		var selection models.PickableSelection
		// Preload Event para validar fechas
		if err := tx.Preload("Event").First(&selection, selectionID).Error; err != nil {
			continue // Si una selección no existe, la saltamos o podríamos retornar error
		}

		// Validar que el evento de la selección pertenezca al torneo actual (seguridad)
		tournamentIDUint := utils.StringToUint(tournamentID)
		if selection.Event.TournamentID != tournamentIDUint {
			tx.Rollback()
			utils.Error(c, http.StatusBadRequest, "La selección #"+utils.UintToString(selection.ID)+" no pertenece a este torneo.", nil)
			return
		}

		// 3. Validar que el evento no haya comenzado
		if time.Now().After(selection.Event.StartTime) {
			tx.Rollback()
			utils.Error(c, http.StatusBadRequest, "El evento '"+selection.Event.Name+"' ya ha comenzado o cerrado", nil)
			return
		}

		// 4. Guardar o Actualizar el Pick
		// Buscamos si ya existe un pick para esta selección por este participante para evitar duplicados
		var existingPick models.UserPick
		err := tx.Where("participant_id = ? AND selection_id = ?", participant.ID, selection.ID).First(&existingPick).Error

		if err == nil {
			// Ya existe, no hacemos nada (o podríamos actualizar si hubiera campos extra)
			savedPicks = append(savedPicks, existingPick)
		} else {
			// Crear nuevo pick
			newPick := models.UserPick{
				ParticipantID: participant.ID,
				SelectionID:   selection.ID,
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
	utils.Success(c, http.StatusCreated, "Predicciones guardadas", savedPicks)
}
