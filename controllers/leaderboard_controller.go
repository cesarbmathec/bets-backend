package controllers

import (
	"net/http"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

// GetTournamentLeaderboard godoc
// @Summary      Ver tabla de clasificación
// @Description  Obtiene la lista de participantes ordenados por puntaje descendente
// @Tags         tournaments
// @Param        id path int true "ID del Torneo"
// @Success      200 {object} utils.Response{data=[]models.TournamentParticipant}
// @Router       /tournaments/{id}/leaderboard [get]
func GetTournamentLeaderboard(c *gin.Context) {
	tournamentID := c.Param("id")
	var participants []models.TournamentParticipant

	// Preload User para mostrar nombres, ordenados por puntos DESC
	if err := config.DB.Preload("User").
		Where("tournament_id = ?", tournamentID).
		Order("total_points desc").
		Find(&participants).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener clasificación", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Tabla de clasificación", participants)
}
