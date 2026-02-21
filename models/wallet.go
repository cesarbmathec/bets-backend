package models

import "time"

type Wallet struct {
	BaseModel
	UserID uint `gorm:"uniqueIndex;not null" json:"user_id"`

	// Balance: Dinero líquido disponible para retirar o apostar
	Balance float64 `gorm:"type:decimal(12,2);default:0;check:balance >= 0" json:"balance"`

	// FrozenBalance: Dinero "en juego" que no se puede retirar ni usar para otras apuestas
	FrozenBalance float64 `gorm:"type:decimal(12,2);default:0" json:"frozen_balance"`

	BonusBalance float64 `gorm:"type:decimal(12,2);default:0" json:"bonus_balance"`
	TokenBalance int     `gorm:"default:0" json:"token_balance"`
	Currency     string  `gorm:"size:10;default:'USD'" json:"currency"`

	// Auditoría de última actualización
	LastTransactionAt *time.Time `json:"last_transaction_at"`
}

// CanAfford verifica si el usuario tiene suficiente saldo (real + bono)
func (w *Wallet) CanAfford(amount float64) bool {
	return (w.Balance + w.BonusBalance) >= amount
}

// TableName define el nombre de la tabla
func (Wallet) TableName() string {
	return "wallets"
}
