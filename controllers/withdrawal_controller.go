package controllers

import (
	"net/http"
	"time"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
)

// Constants for withdrawal limits and verification
const (
	MaxWithdrawalPerDay     float64 = 1000.00
	MaxWithdrawalPerWeek    float64 = 5000.00
	MaxWithdrawalPerMonth   float64 = 20000.00
	VerificationExpiryMins  int     = 30
	MaxVerificationAttempts int     = 3
)

// WithdrawalVerification stores verification attempts in memory
// In production, this should be in Redis or database
var withdrawalVerifications = make(map[uint]struct {
	Code       string
	Attempts   int
	ExpiresAt  time.Time
	Verified   bool
	VerifiedAt *time.Time
})

// CreateWithdrawal godoc
// @Summary      Solicitar retiro
// @Description  Crea una solicitud de retiro con verificación de código
// @Tags         wallet
// @Security     BearerAuth
// @Param        withdrawal body dtos.WithdrawalRequest true "Datos del retiro"
// @Success      200 {object} utils.Response
// @Router       /wallet/withdraw [post]
func CreateWithdrawal(c *gin.Context) {
	userID, _ := c.Get("userID")

	var input dtos.WithdrawalRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar monto mínimo
	if input.Amount < 10.00 {
		utils.Error(c, http.StatusBadRequest, "Monto mínimo de retiro es $10.00", nil)
		return
	}

	// Obtener billetera del usuario
	var wallet models.Wallet
	if err := config.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Billetera no encontrada", nil)
		return
	}

	// Verificar saldo disponible (no se puede usar saldo congelado)
	availableBalance := wallet.Balance + wallet.BonusBalance
	if availableBalance < input.Amount {
		utils.Error(c, http.StatusBadRequest, "Saldo insuficiente", nil)
		return
	}

	// Verificar límites de retiro
	limits, _ := getWithdrawalLimits(userID.(uint), availableBalance)
	if input.Amount > limits.AvailableToday {
		utils.Error(c, http.StatusBadRequest, "Excedes el límite de retiro diario", nil)
		return
	}

	// Verificar que el método de pago exista y pertenezca al usuario
	var paymentMethod models.UserPaymentMethod
	if err := config.DB.Where("id = ? AND user_id = ?", input.PaymentMethodID, userID).First(&paymentMethod).Error; err != nil {
		utils.Error(c, http.StatusBadRequest, "Método de pago no encontrado", nil)
		return
	}

	// Crear solicitud de retiro
	withdrawal := models.Withdrawal{
		UserID:          userID.(uint),
		Amount:          input.Amount,
		PreviousBalance: availableBalance,
		NewBalance:      availableBalance - input.Amount,
		PaymentMethodID: input.PaymentMethodID,
		Status:          "pending",
	}

	// Iniciar transacción
	tx := config.DB.Begin()

	if err := tx.Create(&withdrawal).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusInternalServerError, "Error al crear retiro", nil)
		return
	}

	// Congelar el monto en la billetera
	wallet.FrozenBalance += input.Amount
	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusInternalServerError, "Error al actualizar billetera", nil)
		return
	}

	tx.Commit()

	// Generar código de verificación
	withdrawalVerifications[withdrawal.ID] = struct {
		Code       string
		Attempts   int
		ExpiresAt  time.Time
		Verified   bool
		VerifiedAt *time.Time
	}{
		Code:      withdrawal.WithdrawalCode,
		Attempts:  0,
		ExpiresAt: time.Now().Add(time.Duration(VerificationExpiryMins) * time.Minute),
		Verified:  false,
	}

	// Respuesta con el código de verificación
	response := dtos.WithdrawalWithCodeResponse{
		ID:             withdrawal.ID,
		Amount:         withdrawal.Amount,
		Status:         withdrawal.Status,
		WithdrawalCode: withdrawal.WithdrawalCode,
		Message:        "Código de verificación enviado. Tienes 30 minutos para verificar el retiro.",
		ExpiresIn:      VerificationExpiryMins,
	}

	utils.Success(c, http.StatusOK, "Retiro solicitado. Verifica con el código enviado.", response)
}

// VerifyWithdrawal godoc
// @Summary      Verificar retiro con código
// @Description  Verifica el código de un retiro solicitado
// @Tags         wallet
// @Security     BearerAuth
// @Param        verify body dtos.VerifyWithdrawalRequest true "Código de verificación"
// @Success      200 {object} utils.Response
// @Router       /wallet/withdraw/verify [post]
func VerifyWithdrawal(c *gin.Context) {
	userID, _ := c.Get("userID")

	var input dtos.VerifyWithdrawalRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Código inválido", nil)
		return
	}

	// Buscar el retiro
	var withdrawal models.Withdrawal
	if err := config.DB.Where("id = ? AND user_id = ? AND status = ?", input.WithdrawalID, userID, "pending").First(&withdrawal).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Retiro no encontrado o ya procesado", nil)
		return
	}

	// Verificar código
	verification, exists := withdrawalVerifications[withdrawal.ID]
	if !exists {
		utils.Error(c, http.StatusBadRequest, "Solicitud de verificación expirada", nil)
		return
	}

	// Verificar si ya está verificado
	if verification.Verified {
		utils.Error(c, http.StatusBadRequest, "Este retiro ya ha sido verificado", nil)
		return
	}

	// Verificar si expiró
	if time.Now().After(verification.ExpiresAt) {
		// Expirar el retiro
		rejectWithdrawal(withdrawal.ID, "Código de verificación expirado")
		utils.Error(c, http.StatusBadRequest, "El código ha expirado. Por favor, solicita un nuevo retiro.", nil)
		return
	}

	// Verificar código
	if input.Code != verification.Code {
		verification.Attempts++
		withdrawalVerifications[withdrawal.ID] = verification

		if verification.Attempts >= MaxVerificationAttempts {
			// Bloquear después de 3 intentos fallidos
			rejectWithdrawal(withdrawal.ID, "Demasiados intentos fallidos")
			utils.Error(c, http.StatusTooManyRequests, "Has excedido los intentos permitidos. El retiro ha sido cancelado.", nil)
			return
		}

		remaining := MaxVerificationAttempts - verification.Attempts
		utils.Error(c, http.StatusBadRequest, "Código incorrecto", map[string]interface{}{
			"attempts_remaining": remaining,
		})
		return
	}

	// Código correcto - marcar como verificado
	now := time.Now()
	verification.Verified = true
	verification.VerifiedAt = &now
	withdrawalVerifications[withdrawal.ID] = verification

	// Actualizar retiro
	withdrawal.Verified = true
	withdrawal.VerifiedAt = &now
	config.DB.Save(&withdrawal)

	// Enviar notificación al usuario (en producción, enviar email/SMS)
	// Por ahora solo respondemos éxito

	utils.Success(c, http.StatusOK, "Retiro verificado exitosamente", map[string]interface{}{
		"withdrawal_id": withdrawal.ID,
		"amount":        withdrawal.Amount,
		"status":        "verified",
	})
}

// GetWithdrawalHistory godoc
// @Summary      Historial de retiros
// @Description  Obtiene el historial de retiros del usuario
// @Tags         wallet
// @Security     BearerAuth
// @Success      200 {object} utils.Response
// @Router       /wallet/withdraw/history [get]
func GetWithdrawalHistory(c *gin.Context) {
	userID, _ := c.Get("userID")

	var withdrawals []models.Withdrawal
	if err := config.DB.Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&withdrawals).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener historial", nil)
		return
	}

	// Contar por estado
	pending := 0
	approved := 0
	rejected := 0
	completed := 0

	response := make([]dtos.WithdrawalResponse, len(withdrawals))
	for i, w := range withdrawals {
		// Cargar método de pago
		var paymentMethod models.UserPaymentMethod
		config.DB.First(&paymentMethod, w.PaymentMethodID)

		response[i] = dtos.WithdrawalResponse{
			ID:              w.ID,
			Amount:          w.Amount,
			PreviousBalance: w.PreviousBalance,
			NewBalance:      w.NewBalance,
			Status:          w.Status,
			Verified:        w.Verified,
			VerifiedAt:      w.VerifiedAt,
			RejectedReason:  w.RejectedReason,
			ProcessedAt:     w.ProcessedAt,
			CreatedAt:       w.CreatedAt,
			PaymentMethod: dtos.UserPaymentMethodResponse{
				ID:            paymentMethod.ID,
				Method:        paymentMethod.Method,
				IsDefault:     paymentMethod.IsDefault,
				IsVerified:    paymentMethod.IsVerified,
				PhoneNumber:   paymentMethod.PhoneNumber,
				BankName:      paymentMethod.BankName,
				ZelleEmail:    paymentMethod.ZelleEmail,
				CryptoAddress: paymentMethod.CryptoAddress,
				PaypalEmail:   paymentMethod.PaypalEmail,
				AccountNumber: paymentMethod.AccountNumber,
			},
		}

		switch w.Status {
		case "pending":
			pending++
		case "approved":
			approved++
		case "rejected":
			rejected++
		case "completed":
			completed++
		}
	}

	history := dtos.WithdrawalHistoryResponse{
		Withdrawals: response,
		Total:       len(withdrawals),
		Pending:     pending,
		Approved:    approved,
		Rejected:    rejected,
		Completed:   completed,
	}

	utils.Success(c, http.StatusOK, "Historial de retiros", history)
}

// GetWithdrawalLimits godoc
// @Summary      Límites de retiro
// @Description  Obtiene los límites de retiro del usuario
// @Tags         wallet
// @Security     BearerAuth
// @Success      200 {object} utils.Response
// @Router       /wallet/withdraw/limits [get]
func GetWithdrawalLimits(c *gin.Context) {
	userID, _ := c.Get("userID")

	var wallet models.Wallet
	if err := config.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Billetera no encontrada", nil)
		return
	}

	availableBalance := wallet.Balance + wallet.BonusBalance
	limits, _ := getWithdrawalLimits(userID.(uint), availableBalance)

	utils.Success(c, http.StatusOK, "Límites de retiro", limits)
}

// GetPendingWithdrawal godoc
// @Summary      Obtener retiro pendiente
// @Description  Obtiene el retiro pendiente actual para completar la verificación
// @Tags         wallet
// @Security     BearerAuth
// @Success      200 {object} utils.Response
// @Router       /wallet/withdraw/pending [get]
func GetPendingWithdrawal(c *gin.Context) {
	userID, _ := c.Get("userID")

	var withdrawal models.Withdrawal
	if err := config.DB.Where("user_id = ? AND status = ?", userID, "pending").Order("created_at desc").First(&withdrawal).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "No hay retiros pendientes", nil)
		return
	}

	// Verificar si ya expiró
	if verification, exists := withdrawalVerifications[withdrawal.ID]; exists {
		if time.Now().After(verification.ExpiresAt) && !verification.Verified {
			rejectWithdrawal(withdrawal.ID, "Tiempo de verificación expirado")
			utils.Error(c, http.StatusBadRequest, "El tiempo de verificación ha expirado", nil)
			return
		}
	}

	response := map[string]interface{}{
		"id":                 withdrawal.ID,
		"amount":             withdrawal.Amount,
		"created_at":         withdrawal.CreatedAt,
		"expires_in_minutes": VerificationExpiryMins - int(time.Since(withdrawal.CreatedAt).Minutes()),
		"verified":           withdrawal.Verified,
	}

	utils.Success(c, http.StatusOK, "Retiro pendiente", response)
}

// CancelWithdrawal godoc
// @Summary      Cancelar retiro pendiente
// @Description  Cancela un retiro que aún está pendiente
// @Tags         wallet
// @Security     BearerAuth
// @Param        cancel body dtos.CancelWithdrawalRequest true "ID del retiro"
// @Success      200 {object} utils.Response
// @Router       /wallet/withdraw/cancel [post]
func CancelWithdrawal(c *gin.Context) {
	userID, _ := c.Get("userID")

	var input dtos.CancelWithdrawalRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos inválidos", nil)
		return
	}

	var withdrawal models.Withdrawal
	if err := config.DB.Where("id = ? AND user_id = ? AND status = ?", input.WithdrawalID, userID, "pending").First(&withdrawal).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Retiro no encontrado o ya procesado", nil)
		return
	}

	// Revertir el monto congelado
	tx := config.DB.Begin()

	var wallet models.Wallet
	if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		tx.Rollback()
		utils.Error(c, http.StatusNotFound, "Billetera no encontrada", nil)
		return
	}

	// Descongelar el monto
	wallet.FrozenBalance -= withdrawal.Amount
	tx.Save(&wallet)

	// Marcar retiro como cancelado
	withdrawal.Status = "cancelled"
	tx.Save(&withdrawal)

	tx.Commit()

	// Limpiar verificación
	delete(withdrawalVerifications, withdrawal.ID)

	utils.Success(c, http.StatusOK, "Retiro cancelado exitosamente", nil)
}

// Helper function to reject a withdrawal
func rejectWithdrawal(withdrawalID uint, reason string) {
	tx := config.DB.Begin()

	var withdrawal models.Withdrawal
	if err := tx.First(&withdrawal, withdrawalID).Error; err != nil {
		tx.Rollback()
		return
	}

	// Revertir el monto congelado
	var wallet models.Wallet
	if err := tx.Where("user_id = ?", withdrawal.UserID).First(&wallet).Error; err == nil {
		wallet.FrozenBalance -= withdrawal.Amount
		tx.Save(&wallet)
	}

	// Marcar como rechazado
	withdrawal.Status = "rejected"
	withdrawal.RejectedReason = reason
	tx.Save(&withdrawal)

	tx.Commit()

	// Limpiar verificación
	delete(withdrawalVerifications, withdrawalID)
}

// Helper function to get withdrawal limits
func getWithdrawalLimits(userID uint, availableBalance float64) (dtos.WithdrawalLimitResponse, bool) {
	// Calcular usado hoy
	today := time.Now().Truncate(24 * time.Hour)
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	var usedToday, usedThisWeek, usedThisMonth float64

	var withdrawals []models.Withdrawal
	config.DB.Where("user_id = ? AND status IN (?, ?) AND created_at >= ?", userID, "completed", "approved", today).Find(&withdrawals)
	for _, w := range withdrawals {
		usedToday += w.Amount
	}

	config.DB.Where("user_id = ? AND status IN (?, ?) AND created_at >= ?", userID, "completed", "approved", weekAgo).Find(&withdrawals)
	for _, w := range withdrawals {
		usedThisWeek += w.Amount
	}

	config.DB.Where("user_id = ? AND status IN (?, ?) AND created_at >= ?", userID, "completed", "approved", monthAgo).Find(&withdrawals)
	for _, w := range withdrawals {
		usedThisMonth += w.Amount
	}

	limits := dtos.WithdrawalLimitResponse{
		MaxWithdrawalPerDay:   MaxWithdrawalPerDay,
		MaxWithdrawalPerWeek:  MaxWithdrawalPerWeek,
		MaxWithdrawalPerMonth: MaxWithdrawalPerMonth,
		UsedToday:             usedToday,
		UsedThisWeek:          usedThisWeek,
		UsedThisMonth:         usedThisMonth,
		AvailableToday:        MaxWithdrawalPerDay - usedToday,
		AvailableThisWeek:     MaxWithdrawalPerWeek - usedThisWeek,
		AvailableThisMonth:    MaxWithdrawalPerMonth - usedThisMonth,
	}

	// Ajustar si el balance es menor
	if limits.AvailableToday > availableBalance {
		limits.AvailableToday = availableBalance
	}
	if limits.AvailableThisWeek > availableBalance {
		limits.AvailableThisWeek = availableBalance
	}
	if limits.AvailableThisMonth > availableBalance {
		limits.AvailableThisMonth = availableBalance
	}

	return limits, true
}
