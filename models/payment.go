package models

import (
	"time"
)

type Payment struct {
	BaseModel
	UserID          uint       `gorm:"not null" json:"user_id"`
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount" binding:"required,gt=0"`
	Type            string     `gorm:"size:20;not null" json:"type" binding:"required,oneof=in out"` // in: depósito, out: retiro
	Method          string     `gorm:"size:50;not null" json:"method"`                               // "pago_movil", "zelle", "binance"
	ReferenceNumber string     `gorm:"size:100;uniqueIndex" json:"reference_number" binding:"required"`
	BankName        string     `gorm:"size:100" json:"bank_name"`
	Status          string     `gorm:"size:20;default:'pending'" json:"status"` // pending, approved, rejected
	Notes           string     `gorm:"type:text" json:"notes"`
	VerifiedBy      *uint      `json:"verified_by"`
	VerifiedAt      *time.Time `json:"verified_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Payment) TableName() string {
	return "payments"
}
