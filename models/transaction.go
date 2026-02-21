package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	BaseModel
	TransactionNumber string  `gorm:"size:50;uniqueIndex;not null" json:"transaction_number"`
	WalletID          uint    `gorm:"not null" json:"wallet_id"`
	Amount            float64 `gorm:"type:decimal(12,2);not null" json:"amount"`
	PreviousBalance   float64 `gorm:"type:decimal(12,2)" json:"previous_balance"`
	NewBalance        float64 `gorm:"type:decimal(12,2)" json:"new_balance"`
	Type              string  `gorm:"size:30;not null" json:"type"` // "deposit", "withdraw", "bet_payment", "bet_refund", "prize"
	Currency          string  `gorm:"size:10;default:'USD'" json:"currency"`
	Description       string  `gorm:"type:text" json:"description"`
	ReferenceID       *uint   `json:"reference_id"`   // ID de la Entry o del Payment relacionado
	ReferenceType     string  `json:"reference_type"` // "entries", "payments"
	Status            string  `gorm:"size:20;default:'completed'" json:"status"`

	// Relaciones
	Wallet Wallet `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
}

func (Transaction) TableName() string {
	return "transactions"
}

// BeforeCreate genera un número de transacción único
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.TransactionNumber == "" {
		now := time.Now()
		t.TransactionNumber = fmt.Sprintf("TRX-%s-%d", now.Format("20060102150405"), t.WalletID)
	}
	return nil
}
