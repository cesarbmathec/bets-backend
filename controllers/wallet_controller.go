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

// GetBalance godoc
// @Summary      Consultar saldo de la billetera
// @Description  Obtiene el balance detallado del usuario (incluye saldo congelado y bonos)
// @Tags         wallet
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dtos.WalletResponse}
// @Router       /wallet/balance [get]
func GetBalance(c *gin.Context) {
	userID, _ := c.Get("userID")
	var wallet models.Wallet

	if err := config.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Billetera no encontrada", nil)
		return
	}

	// Mapeamos al DTO de forma explícita
	response := dtos.WalletResponse{
		Balance:        wallet.Balance,
		Bonus:          wallet.BonusBalance,
		Frozen:         wallet.FrozenBalance,
		TotalAvailable: wallet.Balance + wallet.BonusBalance,
		Currency:       wallet.Currency,
	}

	utils.Success(c, http.StatusOK, "Saldo obtenido", response)
}

// DepositMoney godoc
// @Summary      Depositar dinero (Simulación)
// @Description  Añade saldo a la billetera del usuario
// @Tags         wallet
// @Security     BearerAuth
// @Param        amount body object true "Monto a depositar"
// @Success      200 {object} utils.Response
// @Router       /wallet/deposit [post]
func DepositMoney(c *gin.Context) {
	var input struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Monto inválido", err.Error())
		return
	}

	userID, _ := c.Get("userID")

	// Usamos una transacción de BD para asegurar integridad
	tx := config.DB.Begin()

	var wallet models.Wallet
	if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusNotFound, "Billetera no encontrada", nil)
		return
	}

	previousBalance := wallet.Balance
	wallet.Balance += input.Amount
	tx.Save(&wallet)

	// Registrar la transacción
	transaction := models.Transaction{
		WalletID:        wallet.ID,
		Amount:          input.Amount,
		PreviousBalance: previousBalance,
		NewBalance:      wallet.Balance,
		Type:            "deposit",
		Description:     "Depósito manual de prueba",
		Status:          "completed",
	}
	tx.Create(&transaction)

	tx.Commit()
	utils.Success(c, http.StatusOK, "Depósito exitoso", wallet)
}

// GetTransactionHistory godoc
// @Summary      Historial de movimientos
// @Description  Obtiene la lista de todas las transacciones (depósitos, apuestas, premios) del usuario
// @Tags         wallet
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=[]dtos.TransactionResponse}
// @Router       /wallet/history [get]
func GetTransactionHistory(c *gin.Context) {
	userID, _ := c.Get("userID")
	var wallet models.Wallet

	// 1. Buscamos la billetera del usuario
	if err := config.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Billetera no encontrada", nil)
		return
	}

	// 2. Buscamos las transacciones ordenadas por la más reciente
	var transactions []models.Transaction
	if err := config.DB.Where("wallet_id = ?", wallet.ID).Order("created_at desc").Find(&transactions).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener transacciones", nil)
		return
	}

	// 3. Mapeamos a los DTOs
	response := make([]dtos.TransactionResponse, len(transactions))
	for i, t := range transactions {
		response[i] = dtos.TransactionResponse{
			TransactionNumber: t.TransactionNumber,
			Amount:            t.Amount,
			Type:              t.Type,
			Status:            t.Status,
			Description:       t.Description,
			CreatedAt:         t.CreatedAt,
			NewBalance:        t.NewBalance,
		}
	}

	utils.Success(c, http.StatusOK, "Historial obtenido", response)
}

// GetUserStatistics godoc
// @Summary      Estadísticas financieras del usuario
// @Description  Devuelve conteos de depósitos/retiros y montos totales ganados/gastados
// @Tags         wallet
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dtos.UserStatsResponse}
// @Router       /wallet/statistics [get]
func GetUserStatistics(c *gin.Context) {
	userID, _ := c.Get("userID")
	var wallet models.Wallet

	if err := config.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Billetera no encontrada", nil)
		return
	}

	var stats dtos.UserStatsResponse

	// Contar Depósitos
	config.DB.Model(&models.Transaction{}).
		Where("wallet_id = ? AND type = ?", wallet.ID, "deposit").
		Count(&stats.TotalDepositsCount)

	// Contar Retiros
	config.DB.Model(&models.Transaction{}).
		Where("wallet_id = ? AND type = ?", wallet.ID, "withdraw").
		Count(&stats.TotalWithdrawalsCount)

	// Sumar Ganancias (Premios)
	// Usamos COALESCE para evitar NULL si no hay registros
	config.DB.Model(&models.Transaction{}).
		Where("wallet_id = ? AND type = ?", wallet.ID, "prize").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalWinnings)

	// Sumar Gastos (Inscripciones)
	// Nota: En la BD se guardan como negativo, aquí lo mostramos como valor absoluto positivo para "Gasto"
	var totalEntries float64
	config.DB.Model(&models.Transaction{}).
		Where("wallet_id = ? AND type = ?", wallet.ID, "tournament_entry").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalEntries)

	// Convertimos a positivo si está negativo
	if totalEntries < 0 {
		stats.TotalSpent = -totalEntries
	} else {
		stats.TotalSpent = totalEntries
	}

	utils.Success(c, http.StatusOK, "Estadísticas obtenidas", stats)
}
