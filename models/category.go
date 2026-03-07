package models

import (
	"database/sql/driver"
	"encoding/json"
)

// Category representa una categoría de torneo
type Category struct {
	BaseModel
	Name        string `gorm:"size:50;not null;uniqueIndex" json:"name" binding:"required"`
	Slug        string `gorm:"size:60;uniqueIndex;not null" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	Icon        string `gorm:"size:50" json:"icon"`           // Icono representativo (ej: "trophy", "horse")
	Color       string `gorm:"size:7" json:"color"`           // Color hexadecimal (ej: "#FF5733")
	IsActive    bool   `gorm:"default:true" json:"is_active"` // Si la categoría está disponible
	SortOrder   int    `gorm:"default:0" json:"sort_order"`   // Orden de显示
}

// TournamentSelectionType define los tipos de selección disponibles para una categoría
type CategorySelectionType struct {
	BaseModel
	CategoryID    uint     `gorm:"not null;index" json:"category_id"`
	Category      Category `gorm:"foreignKey:CategoryID" json:"-"`
	SelectionType string   `gorm:"size:20;not null" json:"selection_type"` // "macho", "hembra", "alta", "baja", "runline"
	DisplayName   string   `gorm:"size:50" json:"display_name"`            // Nombre a mostrar (ej: "Macho (Favorito)")
	Description   string   `gorm:"type:text" json:"description"`
	IsRequired    bool     `gorm:"default:false" json:"is_required"` // Si es obligatorio para esta categoría
}

// CategorySettingsJSON para guardar configuración adicional en JSON
type CategorySettingsJSON struct {
	MinSelectionsPerSession int  `json:"min_selections_per_session"`
	MaxSelectionsPerSession int  `json:"max_selections_per_session"`
	AllowMultipleEvents     bool `json:"allow_multiple_events"`
}

// Scan implementa el scanner para JSON
func (cs *CategorySettingsJSON) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return json.Unmarshal(value.([]byte), cs)
}

// Value implementa el valuer para JSON
func (cs CategorySettingsJSON) Value() (driver.Value, error) {
	return json.Marshal(cs)
}

func (Category) TableName() string {
	return "categories"
}

func (CategorySelectionType) TableName() string {
	return "category_selection_types"
}
