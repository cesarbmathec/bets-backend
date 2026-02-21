package controllers

import (
	"fmt"
	"net/http"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"

	_ "github.com/cesarbmathec/bets-backend/docs"
)

// CreateTournament godoc
// @Summary      Crear un nuevo torneo
// @Description  Permite a un administrador crear un torneo (Hípica, Fútbol, etc.) con configuración de sesiones y selecciones
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body dtos.CreateTournamentRequest true "Datos del torneo"
// @Success      201 {object} utils.Response{data=models.Tournament} "Torneo creado exitosamente"
// @Failure      400 {object} utils.Response "Datos inválidos"
// @Failure      401 {object} utils.Response "No autorizado"
// @Failure      403 {object} utils.Response "Se requiere rol de administrador"
// @Router       /admin/tournaments [post]
// @Security     BearerAuth
// @example request -json {"name": "Quiniela Semanal", "description": "Torneo de pronosticos", "category": "Futbol", "start_date": "2026-02-23T00:00:00Z", "end_date": "2026-03-01T23:59:59Z", "entry_fee": 10.00, "prize_bonus": 10.00, "settings": {"selections_per_session": 5, "required_selection_types": ["macho", "hembra", "alta", "baja", "runline"], "total_sessions": 5, "prize_distribution": [0.70, 0.20, 0.10]}}
// @example response -json {"success": true, "message": "Torneo creado con exito", "data": {"id": 1, "name": "Quiniela Semanal", "slug": "quiniela-semanal", "status": "open"}}
func CreateTournament(c *gin.Context) {
	var input dtos.CreateTournamentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Extraemos el ID del admin
	userID, _ := c.Get("userID")

	// Mapeo del DTO de settings al modelo de settings
	settings := models.TournamentSettings{
		PrizeDistribution:      input.Settings.PrizeDistribution,
		SelectionsPerSession:   input.Settings.SelectionsPerSession,
		HorseRacingPoints:      input.Settings.HorseRacingPoints,
		RequiredSelectionTypes: input.Settings.RequiredSelectionTypes,
		TotalSessions:          input.Settings.TotalSessions,
	}

	tournament := models.Tournament{
		Name:            input.Name,
		Description:     input.Description,
		Category:        input.Category,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		EntryFee:        input.EntryFee,
		EntryFeeTokens:  input.EntryFeeTokens,
		PrizeBonus:      input.PrizeBonus,
		AdminFeePercent: input.AdminFeePercent,
		Settings:        settings,
		CreatedBy:       userID.(uint),
		Status:          "open",
	}

	if err := config.DB.Create(&tournament).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "No se pudo crear el torneo", nil)
		return
	}

	utils.Success(c, http.StatusCreated, "Torneo creado con éxito", tournament)
}

// GetTournaments godoc
// @Summary      Listar torneos
// @Description  Obtiene todos los torneos disponibles en el sistema
// @Tags         tournaments
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.Tournament} "Lista de torneos"
// @Router       /tournaments [get]
// @example response -json {"success": true, "message": "Lista de torneos", "data": [{"id": 1, "name": "Quiniela Semanal", "category": "Futbol", "status": "open", "entry_fee": 10.00}]}
func GetTournaments(c *gin.Context) {
	var tournaments []models.Tournament
	config.DB.Find(&tournaments)
	utils.Success(c, http.StatusOK, "Lista de torneos", tournaments)
}

// GetTournamentByID godoc
// @Summary      Ver detalle de torneo por ID
// @Description  Obtiene la información completa de un torneo usando su ID numérico
// @Tags         tournaments
// @Param        id path int true "ID del Torneo"
// @Produce      json
// @Success      200 {object} utils.Response{data=models.Tournament} "Torneo encontrado"
// @Failure      404 {object} utils.Response "Torneo no encontrado"
// @Router       /tournaments/id/{id} [get]
func GetTournamentByID(c *gin.Context) {
	id := c.Param("id")
	var tournament models.Tournament

	if err := config.DB.First(&tournament, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Torneo no encontrado por ID", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Torneo encontrado", tournament)
}

// UpdateTournamentStatus godoc
// @Summary      Actualizar estado o finalizar torneo
// @Description  Cambia el estado del torneo (open, closed, finished). Al cambiar a finished se distribuyen los premios.
// @Tags         admin
// @Param        id path int true "ID del Torneo"
// @Param        request body dtos.UpdateStatusRequest true "Nuevo estado"
// @Success      200 {object} utils.Response "Estado actualizado"
// @Failure      400 {object} utils.Response "Estado inválido"
// @Failure      404 {object} utils.Response "Torneo no encontrado"
// @Failure      403 {object} utils.Response "Se requiere rol de administrador"
// @Router       /admin/tournaments/{id}/status [patch]
// @Security     BearerAuth
// @example request -json {"status": "finished"}
func UpdateTournamentStatus(c *gin.Context) {
	id := c.Param("id")
	var input dtos.UpdateStatusRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	var tournament models.Tournament
	if err := config.DB.First(&tournament, id).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Torneo no encontrado", nil)
		return
	}

	// Lógica de Finalización y Reparto de Premios
	if input.Status == "finished" && tournament.Status != "finished" {
		tx := config.DB.Begin()

		// 1. Obtener ganadores (Ranking) según la cantidad de premios definidos
		var winners []models.TournamentParticipant
		limit := len(tournament.Settings.PrizeDistribution)

		if limit > 0 {
			if err := tx.Where("tournament_id = ?", tournament.ID).
				Order("total_points desc").
				Limit(limit).
				Find(&winners).Error; err != nil {
				tx.Rollback()
				utils.Error(c, http.StatusInternalServerError, "Error al calcular ganadores", err.Error())
				return
			}

			// 2. Calcular Total a Repartir (PrizePool + Bono)
			totalPot := tournament.PrizePool + tournament.PrizeBonus

			// 3. Repartir premios
			for i, winner := range winners {
				if i >= len(tournament.Settings.PrizeDistribution) {
					break
				}
				percentage := tournament.Settings.PrizeDistribution[i]
				prizeAmount := totalPot * percentage

				if prizeAmount > 0 {
					// Actualizar Billetera del Ganador
					var wallet models.Wallet
					if err := tx.Where("user_id = ?", winner.UserID).First(&wallet).Error; err != nil {
						tx.Rollback()
						utils.Error(c, http.StatusInternalServerError, "Error buscando billetera de ganador", nil)
						return
					}

					wallet.Balance += prizeAmount
					if err := tx.Save(&wallet).Error; err != nil {
						tx.Rollback()
						utils.Error(c, http.StatusInternalServerError, "Error depositando premio", nil)
						return
					}

					// Registrar Transacción
					trx := models.Transaction{
						WalletID:    wallet.ID,
						Amount:      prizeAmount,
						Type:        "prize",
						Description: fmt.Sprintf("Premio torneo: %s (Posición %d)", tournament.Name, i+1),
						Status:      "completed",
						NewBalance:  wallet.Balance,
					}
					tx.Create(&trx)
				}
			}
		}
		tournament.Status = "finished"
		tx.Save(&tournament)
		tx.Commit()
	} else {
		// Cambio de estado normal (ej: open -> closed)
		tournament.Status = input.Status
		config.DB.Save(&tournament)
	}

	utils.Success(c, http.StatusOK, "Estado actualizado y premios procesados (si aplica)", tournament)
}

// GetTournamentBySlug godoc
// @Summary      Ver detalle de torneo por Slug
// @Description  Obtiene la información completa de un torneo usando su Slug único
// @Tags         tournaments
// @Param        slug path string true "Slug del Torneo"
// @Produce      json
// @Success      200 {object} utils.Response{data=models.Tournament} "Torneo encontrado"
// @Failure      404 {object} utils.Response "Torneo no encontrado"
// @Router       /tournaments/s/{slug} [get]
func GetTournamentBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var tournament models.Tournament

	if err := config.DB.Where("slug = ?", slug).First(&tournament).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Torneo no encontrado por slug", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Torneo encontrado", tournament)
}

// GetTournamentEvents godoc
// @Summary      Listar eventos y opciones de apuesta de un torneo (Cartilla)
// @Description  Obtiene todos los eventos del torneo con sus competidores y selecciones disponibles (Alta, Baja, etc.)
// @Tags         tournaments
// @Param        id path int true "ID del Torneo"
// @Produce      json
// @Success      200 {object} utils.Response{data=[]models.Event} "Eventos del torneo"
// @Failure      404 {object} utils.Response "Torneo no encontrado"
// @Router       /tournaments/{id}/events [get]
func GetTournamentEvents(c *gin.Context) {
	id := c.Param("id")
	var events []models.Event

	// Preload de Competitors y PickableSelections es vital para mostrar la "Cartilla" completa
	if err := config.DB.Preload("Competitors").Preload("PickableSelections").
		Where("tournament_id = ?", id).
		Order("start_time asc").
		Find(&events).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener eventos", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Eventos y selecciones del torneo", events)
}

/*
// GetTournamentSessions godoc
// @Summary      Listar sesiones de un torneo
// @Description  Obtiene todas las sesiones de un torneo con información de eventos y predicciones
// @Tags         sessions
// @Produce      json
// @Param        id path int true "ID del Torneo"
// @Success      200 {object} utils.Response{data=[]models.Session}
// @Router       /tournaments/{id}/sessions [get]
func GetTournamentSessions(c *gin.Context) {
	id := c.Param("id")
	var sessions []models.Session

	if err := config.DB.Preload("Events").Where("tournament_id = ?", id).Order("session_number asc").Find(&sessions).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener sesiones", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Sesiones del torneo", sessions)
}
*/

// GetTournamentLeaderboard ya está definido en leaderboard_controller.go
