package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// TournamentSettings define las reglas y premios específicos del torneo.
type TournamentSettings struct {
	PrizeDistribution    []float64 `json:"prize_distribution"`     // Ej: [0.7, 0.2, 0.1] para 1ro, 2do, 3ro
	SelectionsPerSession int       `json:"selections_per_session"` // Cantidad de selecciones por sesión/día
	HorseRacingPoints    []int     `json:"horse_racing_points"`    // Ej: [10, 5, 3] para 1ro, 2do, 3ro en carreras

	// Configuración de tipos de selección requeridos por posición
	// Ej: ["macho", "hembra", "alta", "baja", "runline"]
	// Si está vacío, el usuario puede elegir cualquier tipo
	RequiredSelectionTypes []string `json:"required_selection_types"`

	// Cantidad total de sesiones que tendrá el torneo
	TotalSessions int `json:"total_sessions"`
}

// Implementación para guardar JSON en Gorm (MySQL/Postgres)
func (ts *TournamentSettings) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), ts)
}

func (ts TournamentSettings) Value() (driver.Value, error) {
	return json.Marshal(ts)
}

type Tournament struct {
	BaseModel
	Name        string    `gorm:"size:100;not null;index" json:"name" binding:"required"`
	Slug        string    `gorm:"size:120;uniqueIndex;not null" json:"slug"` // URL amigable
	Description string    `gorm:"type:text" json:"description"`
	Category    string    `gorm:"size:50;not null;index" json:"category"` // "Hipica", "Futbol"
	Status      string    `gorm:"size:20;not null;default:'open';index" json:"status"`
	StartDate   time.Time `gorm:"not null" json:"start_date"`
	EndDate     time.Time `gorm:"not null" json:"end_date"`

	// Campos Financieros
	EntryFee        float64 `gorm:"type:decimal(12,2);not null" json:"entry_fee"`            // Costo de inscripción
	EntryFeeTokens  int     `gorm:"default:0" json:"entry_fee_tokens"`                       // Costo de inscripción en Tokens
	PrizePool       float64 `gorm:"type:decimal(12,2);default:0" json:"prize_pool"`          // Dinero acumulado
	PrizeBonus      float64 `gorm:"type:decimal(12,2);default:0" json:"prize_bonus"`         // Dinero de bono agregado por la casa
	AdminFeePercent float64 `gorm:"type:decimal(5,2);default:10.0" json:"admin_fee_percent"` // % de comisión para la casa

	// Configuración dinámica (JSON)
	Settings TournamentSettings `gorm:"type:json" json:"settings"`

	CreatedBy    uint                    `json:"created_by"`
	Creator      User                    `gorm:"foreignKey:CreatedBy" json:"-"`
	Events       []Event                 `gorm:"foreignKey:TournamentID" json:"events,omitempty"`
	Participants []TournamentParticipant `gorm:"foreignKey:TournamentID" json:"participants,omitempty"`
}

// Hook BeforeSave: Se ejecuta antes de Crear o Actualizar
func (t *Tournament) BeforeSave(tx *gorm.DB) (err error) {
	// 1. Generar Slug automáticamente si el nombre cambió o es nuevo
	if t.Name != "" {
		t.Slug = slug.Make(t.Name)
	}

	// 2. Validación de estado
	if t.Status == "finished" && t.PrizePool == 0 {
		// Se podrían añadir más validaciones, como que todos los eventos estén 'completed'
	}

	// 3. Validación de fechas lógica
	if t.EndDate.Before(t.StartDate) {
		return errors.New("la fecha de finalización no puede ser anterior a la de inicio")
	}

	return nil
}

func (Tournament) TableName() string {
	return "tournaments"
}
