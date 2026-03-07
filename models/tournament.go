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
	// Distribución de premios (porcentajes que suman 100)
	// Ej: [0.7, 0.2, 0.1] para 1ro(70%), 2do(20%), 3ro(10%)
	PrizeDistribution []float64 `json:"prize_distribution"`

	// Cantidad de selecciones que el usuario debe hacer por sesión
	SelectionsPerSession int `json:"selections_per_session"`

	// Puntos para carreras de caballos por posición
	// Ej: [10, 5, 3] para 1ro(10pts), 2do(5pts), 3ro(3pts)
	HorseRacingPoints []int `json:"horse_racing_points"`

	// Tipos de selección requeridos por posición en cada sesión
	// Ej: ["macho", "hembra", "alta", "baja"]
	// Si está vacío, el usuario puede elegir cualquier tipo (elección libre)
	RequiredSelectionTypes []string `json:"required_selection_types"`

	// Puntos otorgados por cada tipo de selección
	// Ej: map[string]int{"macho": 3, "hembra": 5, "alta": 3, "baja": 3, "empate": 1}
	PointsBySelectionType map[string]int `json:"points_by_selection_type"`

	// Puntos extras por acertar todas las selecciones de una sesión
	ExtraPointsForPerfectSession int `json:"extra_points_for_perfect_session"`

	// Cantidad total de sesiones que tendrá el torneo
	TotalSessions int `json:"total_sessions"`

	// Es elección libre? (true = puede elegir cualquier tipo de selección)
	FreeSelection bool `json:"free_selection"`

	// Categoría del deporte: "futbol", "beisbol", "caballos", "basquet", etc.
	SportCategory string `json:"sport_category"`
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

	// Límite de participantes (0 = sin límite)
	MaxParticipants int `gorm:"default:0" json:"max_participants"`

	// Campos Financieros
	EntryFee        float64 `gorm:"type:decimal(12,2);not null" json:"entry_fee"`            // Costo de inscripción
	EntryFeeTokens  int     `gorm:"default:0" json:"entry_fee_tokens"`                       // Costo de inscripción en Tokens
	PrizePool       float64 `gorm:"type:decimal(12,2);default:0" json:"prize_pool"`          // Dinero acumulado
	PrizeBonus      float64 `gorm:"type:decimal(12,2);default:0" json:"prize_bonus"`         // Dinero de bono agregado por la casa
	AdminFeePercent float64 `gorm:"type:decimal(5,2);default:10.0" json:"admin_fee_percent"` // % de comisión para la casa

	// Configuración dinámica (JSON)
	Settings TournamentSettings `gorm:"type:json" json:"settings"`

	CreatedBy uint `json:"created_by"`
	Creator   User `gorm:"foreignKey:CreatedBy" json:"-"`
	// La relación con eventos es a través de TournamentEvent (tabla intermedia)
	Events       []TournamentEvent       `gorm:"foreignKey:TournamentID" json:"events,omitempty"`
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
