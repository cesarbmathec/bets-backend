package controllers

import (
	"net/http"

	"github.com/cesarbmathec/bets-backend/config"
	"github.com/cesarbmathec/bets-backend/dtos"
	"github.com/cesarbmathec/bets-backend/models"
	"github.com/cesarbmathec/bets-backend/utils"
	"github.com/gin-gonic/gin"
)

// GetPaymentMethods godoc
// @Summary     Obtener métodos de pago del usuario
// @Description Devuelve todos los métodos de pago registrados por el usuario
// @Tags        payment-methods
// @Produce     json
// @Security 	 BearerAuth
// @Success     200 {object} utils.Response{data=[]dtos.UserPaymentMethodResponse} "Métodos de pago"
// @Router      /api/v1/payment-methods [get]
func GetPaymentMethods(c *gin.Context) {
	userID := c.GetUint("userID")

	var methods []models.UserPaymentMethod
	db := config.GetDB()

	if err := db.Where("user_id = ?", userID).Find(&methods).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al obtener métodos de pago", nil)
		return
	}

	var response []dtos.UserPaymentMethodResponse
	for _, m := range methods {
		response = append(response, dtos.UserPaymentMethodResponse{
			ID:            m.ID,
			Method:        m.Method,
			IsDefault:     m.IsDefault,
			IsVerified:    m.IsVerified,
			PhoneNumber:   m.PhoneNumber,
			BankName:      m.BankName,
			ZelleEmail:    m.ZelleEmail,
			ZelleName:     m.ZelleName,
			CryptoAddress: m.CryptoAddress,
			CryptoNetwork: m.CryptoNetwork,
			PaypalEmail:   m.PaypalEmail,
			AccountNumber: m.AccountNumber,
		})
	}

	utils.Success(c, http.StatusOK, "Métodos de pago", response)
}

// CreatePaymentMethod godoc
// @Summary     Agregar método de pago
// @Description Agrega un nuevo método de pago pararetiros
// @Tags        payment-methods
// @Accept      json
// @Produce     json
// @Param       request body dtos.UserPaymentMethodRequest true "Datos del método de pago"
// @Security 	 BearerAuth
// @Success     201 {object} utils.Response{data=dtos.UserPaymentMethodResponse} "Método de pago creado"
// @Router      /api/v1/payment-methods [post]
func CreatePaymentMethod(c *gin.Context) {
	userID := c.GetUint("userID")

	var input dtos.UserPaymentMethodRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	db := config.GetDB()

	// Si es默认, quitar默认de otros
	if input.IsDefault {
		db.Model(&models.UserPaymentMethod{}).
			Where("user_id = ?", userID).
			Update("is_default", false)
	}

	method := models.UserPaymentMethod{
		UserID: userID,
		Method: input.Method,

		// Pago Móvil
		PhoneNumber: input.PhoneNumber,
		BankName:    input.BankName,
		BankAccount: input.BankAccount,

		// Zelle
		ZelleEmail: input.ZelleEmail,
		ZelleName:  input.ZelleName,

		// Crypto
		CryptoAddress: input.CryptoAddress,
		CryptoNetwork: input.CryptoNetwork,
		CryptoEmail:   input.CryptoEmail,

		// PayPal
		PaypalEmail: input.PaypalEmail,

		// Banco
		AccountNumber: input.AccountNumber,
		AccountType:   input.AccountType,
		CLABE:         input.CLABE,
		SwiftCode:     input.SwiftCode,

		IsDefault: input.IsDefault,
	}

	if err := db.Create(&method).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al guardar método de pago", nil)
		return
	}

	response := dtos.UserPaymentMethodResponse{
		ID:            method.ID,
		Method:        method.Method,
		IsDefault:     method.IsDefault,
		IsVerified:    method.IsVerified,
		PhoneNumber:   method.PhoneNumber,
		BankName:      method.BankName,
		ZelleEmail:    method.ZelleEmail,
		ZelleName:     method.ZelleName,
		CryptoAddress: method.CryptoAddress,
		CryptoNetwork: method.CryptoNetwork,
		PaypalEmail:   method.PaypalEmail,
		AccountNumber: method.AccountNumber,
	}

	utils.Success(c, http.StatusCreated, "Método de pago agregado", response)
}

// DeletePaymentMethod godoc
// @Summary     Eliminar método de pago
// @Description Elimina un método de pago del usuario
// @Tags        payment-methods
// @Produce     json
// @Param       id path int true "ID del método de pago"
// @Security 	 BearerAuth
// @Success     200 {object} utils.Response "Método de pago eliminado"
// @Router      /api/v1/payment-methods/{id} [delete]
func DeletePaymentMethod(c *gin.Context) {
	userID := c.GetUint("userID")
	methodID := c.Param("id")

	db := config.GetDB()

	var method models.UserPaymentMethod
	if err := db.Where("id = ? AND user_id = ?", methodID, userID).First(&method).Error; err != nil {
		utils.Error(c, http.StatusNotFound, "Método de pago no encontrado", nil)
		return
	}

	if err := db.Delete(&method).Error; err != nil {
		utils.Error(c, http.StatusInternalServerError, "Error al eliminar método de pago", nil)
		return
	}

	utils.Success(c, http.StatusOK, "Método de pago eliminado", nil)
}
