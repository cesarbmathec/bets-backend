package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Withdrawal struct {
	BaseModel
	UserID          uint       `gorm:"not null" json:"user_id"`
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	PreviousBalance float64    `gorm:"type:decimal(12,2)" json:"previous_balance"`
	NewBalance      float64    `gorm:"type:decimal(12,2)" json:"new_balance"`
	PaymentMethodID uint       `gorm:"not null" json:"payment_method_id"`
	Status          string     `gorm:"size:20;default:'pending'" json:"status"` // pending, approved, rejected, completed
	WithdrawalCode  string     `gorm:"size:10" json:"withdrawal_code"`          // Código de verificación
	Verified        bool       `gorm:"default:false" json:"verified"`
	VerifiedAt      *time.Time `json:"verified_at"`
	RejectedReason  string     `gorm:"type:text" json:"rejected_reason,omitempty"`
	ProcessedAt     *time.Time `json:"processed_at"`
	ProcessedBy     *uint      `json:"processed_by,omitempty"`

	// Relaciones
	User          User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	PaymentMethod UserPaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"payment_method,omitempty"`
}

func (Withdrawal) TableName() string {
	return "withdrawals"
}

// BeforeCreate genera un código de verificación único
func (w *Withdrawal) BeforeCreate(tx *gorm.DB) error {
	if w.WithdrawalCode == "" {
		// Generar código de 6 dígitos
		w.WithdrawalCode = fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return nil
}
