package dtos

// UserPaymentMethodRequest estructura para crear/actualizar método de pago
type UserPaymentMethodRequest struct {
	Method string `json:"method" binding:"required,oneof=pago_movil zelle binance paypal banco"`

	// Datos de Pago Móvil
	PhoneNumber string `json:"phone_number,omitempty"`
	BankName    string `json:"bank_name,omitempty"`
	BankAccount string `json:"bank_account,omitempty"`

	// Datos de Zelle
	ZelleEmail string `json:"zelle_email,omitempty"`
	ZelleName  string `json:"zelle_name,omitempty"`

	// Datos de Binance/USDT
	CryptoAddress string `json:"crypto_address,omitempty"`
	CryptoNetwork string `json:"crypto_network,omitempty"`
	CryptoEmail   string `json:"crypto_email,omitempty"`

	// Datos de PayPal
	PaypalEmail string `json:"paypal_email,omitempty"`

	// Datos de Transferencia Bancaria
	AccountNumber string `json:"account_number,omitempty"`
	AccountType   string `json:"account_type,omitempty"`
	CLABE         string `json:"clabe,omitempty"`
	SwiftCode     string `json:"swift_code,omitempty"`

	IsDefault bool `json:"is_default"`
}

// UserPaymentMethodResponse estructura de respuesta para método de pago
type UserPaymentMethodResponse struct {
	ID            uint   `json:"id"`
	Method        string `json:"method"`
	IsDefault     bool   `json:"is_default"`
	IsVerified    bool   `json:"is_verified"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	BankName      string `json:"bank_name,omitempty"`
	ZelleEmail    string `json:"zelle_email,omitempty"`
	ZelleName     string `json:"zelle_name,omitempty"`
	CryptoAddress string `json:"crypto_address,omitempty"`
	CryptoNetwork string `json:"crypto_network,omitempty"`
	PaypalEmail   string `json:"paypal_email,omitempty"`
	AccountNumber string `json:"account_number,omitempty"`
}
