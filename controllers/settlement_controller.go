package controllers

import (
	"net/http"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

// SettleEvent godoc
// @Summary      Liquidar un evento y calcular puntos
// @Description  Establece los resultados finales de un evento, evalúa las selecciones y asigna puntos a los participantes.
// @Tags         admin
// @Security     BearerAuth
// @Param        id path int true "ID del Evento a liquidar"
// @Param        request body dtos.SetEventResultRequest true "Resultados finales de los competidores"
// @Success      200 {object} utils.Response
// @Router       /admin/events/{id}/settle [post]
func SettleEvent(c *gin.Context) {
	eventID := c.Param("id")

	var input dtos.SetEventResultRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos de resultado inválidos", err.Error())
		return
	}

	tx := config.DB.Begin()

	// 1. Cargar Evento y sus relaciones
	var event models.Event
	if err := tx.Preload("Competitors").Preload("PickableSelections").First(&event, eventID).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusNotFound, "Evento no encontrado", nil)
		return
	}

	if event.Status == "completed" {
		tx.Rollback()
		utils.Error(c, http.StatusBadRequest, "Este evento ya ha sido liquidado", nil)
		return
	}

	// 2. Actualizar resultados en los competidores del evento
	competitorResults := make(map[uint]dtos.CompetitorResult)
	totalScore := 0
	for _, res := range input.Results {
		competitorResults[res.CompetitorID] = res
		totalScore += res.FinalScore
		err := tx.Model(&models.EventCompetitor{}).Where("event_id = ? AND id = ?", event.ID, res.CompetitorID).Updates(map[string]interface{}{
			"final_score": res.FinalScore,
			"position":    res.Position,
		}).Error
		if err != nil {
			tx.Rollback()
			utils.Error(c, http.StatusInternalServerError, "Error al actualizar resultados de competidores", err.Error())
			return
		}
	}

	// 3. Evaluar cada 'PickableSelection' del evento
	for _, selection := range event.PickableSelections {
		status := "lost" // Por defecto, una selección es perdedora
		points := 0

		switch selection.SelectionType {
		case "alta":
			if float64(totalScore) > selection.Line {
				status = "won"
				points = selection.PointsForWin
			} else if float64(totalScore) == selection.Line {
				status = "push"
				points = selection.PointsForPush
			}
		case "baja":
			if float64(totalScore) < selection.Line {
				status = "won"
				points = selection.PointsForWin
			} else if float64(totalScore) == selection.Line {
				status = "push"
				points = selection.PointsForPush
			}
		case "macho", "hembra", "ganador": // 'macho' y 'hembra' son terminos para favorito/no favorito
			if selection.CompetitorID != nil {
				if result, ok := competitorResults[*selection.CompetitorID]; ok {
					// Asumimos que el ganador es el que tiene la posición 1 o el score más alto en un vs
					if result.Position == 1 { // Ideal para carreras
						status = "won"
						points = selection.PointsForWin
					}
				}
			}
		case "empate":
			if len(event.Competitors) == 2 {
				c1Score := competitorResults[event.Competitors[0].ID].FinalScore
				c2Score := competitorResults[event.Competitors[1].ID].FinalScore
				if c1Score == c2Score {
					status = "won"
					points = selection.PointsForWin
				}
			}
		}

		// Actualizar el estado de la PickableSelection
		tx.Model(&selection).Update("status", status)

		// 4. Actualizar todos los UserPicks que apuntan a esta selección y sumar los puntos
		if points > 0 {
			if err := updateUserPicksAndPoints(tx, selection.ID, status, points); err != nil {
				tx.Rollback()
				utils.Error(c, http.StatusInternalServerError, "Error al actualizar predicciones de usuarios", err.Error())
				return
			}
		}
	}

	// 5. Marcar el evento como completado
	event.Status = "completed"
	if err := tx.Save(&event).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusInternalServerError, "Error al finalizar el evento", err.Error())
		return
	}

	tx.Commit()
	utils.Success(c, http.StatusOK, "Evento liquidado y puntos asignados correctamente", nil)
}

// updateUserPicksAndPoints actualiza los UserPicks y los puntos totales de los participantes.
func updateUserPicksAndPoints(tx *gorm.DB, selectionID uint, status string, points int) error {
	// Actualizar el estado y los puntos de todos los UserPicks para esta selección
	if err := tx.Model(&models.UserPick{}).
		Where("selection_id = ?", selectionID).
		Updates(map[string]interface{}{"status": status, "awarded_points": points}).Error; err != nil {
		return err
	}

	// Sumar los puntos a los participantes correspondientes.
	return tx.Exec(`
        UPDATE tournament_participants
        SET total_points = total_points + ?
        WHERE id IN (
            SELECT participant_id FROM user_picks WHERE selection_id = ?
        )
    `, points, selectionID).Error
}
