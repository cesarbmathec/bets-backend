package models

import (
	"errors"
	"time"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type Event struct {
	BaseModel
	TournamentID uint       `gorm:"index;not null" json:"tournament_id"`
	SessionID    *uint      `gorm:"index" json:"session_id,omitempty"` // Sesión a la que pertenece este evento
	Session      *Session   `gorm:"foreignKey:SessionID" json:"-"`
	Tournament   Tournament `gorm:"foreignKey:TournamentID" json:"-"`

	Name      string    `gorm:"size:200;not null" json:"name" binding:"required"`
	Slug      string    `gorm:"size:220;uniqueIndex;not null" json:"slug"`
	Order     int       `gorm:"default:0" json:"order"` // Ej: Carrera #1, Carrera #2
	StartTime time.Time `gorm:"not null;index" json:"start_time"`

	// Estados: scheduled (programado), live (en vivo), completed (terminado), cancelled
	Status string `gorm:"size:20;default:'scheduled';index" json:"status"`

	// Información de resultados
	ResultNote string `gorm:"type:text" json:"result_note,omitempty"`

	// Relaciones
	Competitors        []EventCompetitor   `gorm:"foreignKey:EventID" json:"competitors"`
	PickableSelections []PickableSelection `gorm:"foreignKey:EventID" json:"pickable_selections,omitempty"`
}

type EventCompetitor struct {
	BaseModel
	EventID        uint   `gorm:"index" json:"event_id"`
	Name           string `gorm:"size:150;not null" json:"name"`
	AssignedNumber int    `json:"assigned_number,omitempty"` // El número en el dorsal/pista

	// Resultado (se actualiza al liquidar el evento)
	FinalScore  int  `gorm:"default:0" json:"final_score,omitempty"` // Para deportes de equipo
	Position    int  `gorm:"default:0" json:"position,omitempty"`    // Para carreras (1 para el ganador)
	IsScratched bool `gorm:"default:false" json:"is_scratched"`      // Para caballos que se retiran antes
}

// Hook para automatizar el Slug y validaciones
func (e *Event) BeforeSave(tx *gorm.DB) (err error) {
	if e.Name != "" {
		e.Slug = slug.Make(e.Name)
	}

	if e.Status == "completed" && len(e.ResultNote) < 5 {
		return errors.New("debe proporcionar una nota de resultado para completar el evento")
	}
	return nil
}
