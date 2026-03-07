package models

import (
	"gorm.io/gorm"
)

// Competitor es un catálogo global de competidores (caballos, equipos, etc.)
// que pueden usarse en diferentes eventos.
type Competitor struct {
	BaseModel
	Name           string `gorm:"size:150;not null" json:"name"`
	Category       string `gorm:"size:100" json:"category"` // Ej: Caballos, Fútbol, Fórmula 1
	AssignedNumber int    `json:"assigned_number"`          // Número dorsal
	Description    string `gorm:"type:text" json:"description"`
	Status         string `gorm:"size:20;default:active" json:"status"` // active, inactive
}

// Hook para validación
func (c *Competitor) BeforeSave(tx *gorm.DB) error {
	if c.Status == "" {
		c.Status = "active"
	}
	return nil
}

// TableName define el nombre de la tabla
func (Competitor) TableName() string {
	return "competitors"
}
