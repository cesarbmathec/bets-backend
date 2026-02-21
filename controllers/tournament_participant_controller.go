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

// JoinTournament godoc
// @Summary      Inscribirse en un torneo
// @Description  Permite a un usuario unirse a un torneo pagando la entrada (EntryFee)
// @Tags         users
// @Security     BearerAuth
// @Param        id path int true "ID del Torneo"
// @Param        request body dtos.JoinTournamentRequest true "Opciones de pago"
// @Success      201 {object} utils.Response
// @Router       /tournaments/{id}/join [post]
func JoinTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	userID, _ := c.Get("userID")

	var input dtos.JoinTournamentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	tx := config.DB.Begin()

	// 1. Validar Torneo
	var tournament models.Tournament
	if err := tx.First(&tournament, tournamentID).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusNotFound, "Torneo no encontrado", nil)
		return
	}

	if tournament.Status != "open" {
		tx.Rollback()
		utils.Error(c, http.StatusBadRequest, "El torneo no está abierto para inscripciones", nil)
		return
	}

	// 2. Verificar si ya está inscrito
	var existing models.TournamentParticipant
	if err := tx.Where("user_id = ? AND tournament_id = ?", userID, tournament.ID).First(&existing).Error; err == nil {
		tx.Rollback()
		utils.Error(c, http.StatusConflict, "Ya estás inscrito en este torneo", nil)
		return
	}

	// 3. Gestionar Pago (Wallet)
	var wallet models.Wallet
	if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusInternalServerError, "Error al obtener billetera", nil)
		return
	}

	cost := tournament.EntryFee
	currency := "USD"

	// Lógica simple: Si pide pagar con tokens y el torneo tiene costo en tokens > 0
	if input.PayWithTokens && tournament.EntryFeeTokens > 0 {
		if wallet.TokenBalance < tournament.EntryFeeTokens {
			tx.Rollback()
			utils.Error(c, http.StatusBadRequest, "Saldo de tokens insuficiente", nil)
			return
		}
		wallet.TokenBalance -= tournament.EntryFeeTokens
		cost = float64(tournament.EntryFeeTokens)
		currency = "TOKENS"
	} else {
		// Pago con dinero real (Saldo + Bono)
		if !wallet.CanAfford(cost) {
			tx.Rollback()
			utils.Error(c, http.StatusBadRequest, "Saldo insuficiente", nil)
			return
		}
		wallet.Balance -= cost // Simplificación: aquí podrías descontar primero del bono si quisieras
	}

	// Guardar cambios en Wallet
	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusInternalServerError, "Error al procesar el pago", nil)
		return
	}

	// 4. Crear Inscripción
	participant := models.TournamentParticipant{
		UserID:       userID.(uint),
		TournamentID: tournament.ID,
		TotalPoints:  0,
	}

	if err := tx.Create(&participant).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusInternalServerError, "Error al registrar inscripción", nil)
		return
	}

	// 5. Registrar Transacción (Auditoría)
	transaction := models.Transaction{
		WalletID:    wallet.ID,
		Amount:      -cost, // Negativo porque es un gasto
		Type:        "tournament_entry",
		Currency:    currency,
		Description: "Inscripción a torneo: " + tournament.Name,
		Status:      "completed",
		NewBalance:  wallet.Balance,
	}
	tx.Create(&transaction)

	// Actualizar PrizePool del torneo (opcional, si el premio crece con las entradas)
	// tournament.PrizePool += cost * (1 - tournament.AdminFeePercent/100)
	// tx.Save(&tournament)

	tx.Commit()
	utils.Success(c, http.StatusCreated, "Inscripción exitosa", participant)
}
