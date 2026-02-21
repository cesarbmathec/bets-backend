package models

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	BaseModel
	Username string `gorm:"size:100;uniqueIndex;not null" json:"username" binding:"required"`
	Nickname string `gorm:"size:100" json:"nickname"`
	Email    string `gorm:"size:150;uniqueIndex;not null" json:"email" binding:"required,email"`
	Password string `gorm:"not null" json:"-"` // Oculto en JSON
	Role     string `gorm:"size:20;default:'user'" json:"role"`
	IsActive bool   `gorm:"default:true" json:"is_active"`

	// Datos personales
	FullName   string `gorm:"size:200" json:"full_name"`
	Phone      string `gorm:"size:20" json:"phone"`
	DocumentID string `gorm:"size:50" json:"document_id"`

	Wallet                 Wallet                  `gorm:"foreignKey:UserID" json:"wallet"`
	TournamentInscriptions []TournamentParticipant `gorm:"foreignKey:UserID" json:"tournament_inscriptions,omitempty"`
	PaymentMethods         []UserPaymentMethod     `gorm:"foreignKey:UserID" json:"payment_methods,omitempty"`
}

func (u *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}
