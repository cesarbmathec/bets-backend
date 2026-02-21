package models

// UserPaymentMethod almacena los métodos de pago/retiro del usuario
type UserPaymentMethod struct {
	BaseModel
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	Method     string `gorm:"size:30;not null" json:"method" binding:"required"` // "pago_movil", "zelle", "binance", "paypal", "banco"
	IsDefault  bool   `gorm:"default:false" json:"is_default"`
	IsVerified bool   `gorm:"default:false" json:"is_verified"`

	// Datos de Pago Móvil
	PhoneNumber string `gorm:"size:20" json:"phone_number,omitempty"`
	BankName    string `gorm:"size:100" json:"bank_name,omitempty"`
	BankAccount string `gorm:"size:50" json:"bank_account,omitempty"`

	// Datos deZelle
	ZelleEmail string `gorm:"size:150" json:"zelle_email,omitempty"`
	ZelleName  string `gorm:"size:100" json:"zelle_name,omitempty"`

	// Datos de Binance/USDT
	CryptoAddress string `gorm:"size:200" json:"crypto_address,omitempty"`
	CryptoNetwork string `gorm:"size:20" json:"crypto_network,omitempty"` // "BEP20", "TRC20", etc.
	CryptoEmail   string `gorm:"size:150" json:"crypto_email,omitempty"`

	// Datos dePayPal
	PaypalEmail string `gorm:"size:150" json:"paypal_email,omitempty"`

	// Datos de Transferencia Bancaria
	AccountNumber string `gorm:"size:50" json:"account_number,omitempty"`
	AccountType   string `gorm:"size:20" json:"account_type,omitempty"` // "corriente", "ahorro"
	CLABE         string `gorm:"size:50" json:"clabe,omitempty"`
	SwiftCode     string `gorm:"size:20" json:"swift_code,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserPaymentMethod) TableName() string {
	return "user_payment_methods"
}
