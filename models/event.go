package models

import (
	"errors"
	"time"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// Event representa un evento/deporte/partido global que puede ser asignado a diferentes torneos.
// Los eventos son independientes de los torneos y se relacionan a través de la tabla TournamentEvent.
type Event struct {
	BaseModel
	// NO tiene TournamentID directo - la relación es a través de TournamentEvent
	// NO tiene SessionID directo - la relación es a través de TournamentEvent

	Name      string    `gorm:"size:200;not null" json:"name" binding:"required"`
	Slug      string    `gorm:"size:220;uniqueIndex;not null" json:"slug"`
	Order     int       `gorm:"default:0" json:"order"` // Ej: Carrera #1, Partido #1
	StartTime time.Time `gorm:"not null;index" json:"start_time"`

	// Lugar/Venue del evento (estadio, hipódromo, etc.)
	Venue string `gorm:"size:200" json:"venue"`

	// Línea Over/Under (ej: 2.5 goles, 7 carreras)
	// Determina si es Alta o Baja
	Line float64 `gorm:"type:decimal(10,2)" json:"line"`

	// Estados: scheduled (programado), live (en vivo), completed (terminado), cancelled
	Status string `gorm:"size:20;default:'scheduled';index" json:"status"`

	// Información de resultados
	ResultNote string `gorm:"type:text" json:"result_note,omitempty"`

	// Total de score/puntos/carreras del evento (para liquidar alta/baja)
	TotalScore float64 `gorm:"type:decimal(10,2);default:0" json:"total_score"`

	// Relaciones
	Competitors        []EventCompetitor   `gorm:"foreignKey:EventID" json:"competitors"`
	PickableSelections []PickableSelection `gorm:"foreignKey:EventID" json:"pickable_selections,omitempty"`
}

// TournamentEvent define la relación muchos a muchos entre Tournament y Event.
// Permite que un evento sea usado en múltiples torneos.
type TournamentEvent struct {
	BaseModel
	TournamentID uint       `gorm:"index;not null" json:"tournament_id"`
	Tournament   Tournament `gorm:"foreignKey:TournamentID" json:"-"`
	EventID      uint       `gorm:"index;not null" json:"event_id"`
	Event        Event      `gorm:"foreignKey:EventID" json:"-"`
	SessionID    *uint      `gorm:"index" json:"session_id,omitempty"` // Sesión específica en este torneo
	Session      Session    `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Order        int        `gorm:"default:0" json:"order"` // Orden del evento en la sesión
}

func (TournamentEvent) TableName() string {
	return "tournament_events"
}

type EventCompetitor struct {
	BaseModel
	EventID uint  `gorm:"index" json:"event_id"`
	Event   Event `gorm:"foreignKey:EventID" json:"-"`

	// Referencia al competidor global (opcional)
	CompetitorID *uint       `gorm:"index" json:"competitor_id,omitempty"`
	Competitor   *Competitor `gorm:"foreignKey:CompetitorID" json:"competitor,omitempty"`

	Name           string `gorm:"size:150;not null" json:"name"`
	AssignedNumber int    `json:"assigned_number"` // El número en el dorsal/pista

	// Odds del competidor (positivo o negativo)
	// Ej: -300, +400, +150
	Odds int `json:"odds"`

	// Runline (para deportes con puntos)
	// Ej: -1.5, +1.5
	Runline float64 `gorm:"type:decimal(10,2)" json:"runline"`

	// Super Runline
	SuperRunline float64 `gorm:"type:decimal(10,2)" json:"super_runline"`

	// Es favorito? (para determinar Macho/Hembra)
	IsFavorite bool `gorm:"default:false" json:"is_favorite"`

	// Resultado (se actualiza al liquidar el evento)
	FinalScore  int  `gorm:"default:0" json:"final_score,omitempty"` // Para deportes de equipo
	Position    int  `gorm:"default:0" json:"position,omitempty"`    // Para carreras (1 para el ganador)
	IsScratched bool `gorm:"default:false" json:"is_scratched"`      // Para caballos que se retiran antes

	// Resultados adicionales
	ScoredFirst       bool `gorm:"default:false" json:"scored_first"`        // Marcó primero
	ScoredFirstHalf   bool `gorm:"default:false" json:"scored_first_half"`   // Marcó en primer tiempo/cuartos
	ScoredSecondHalf  bool `gorm:"default:false" json:"scored_second_half"`  // Marcó en segundo tiempo
	ScoredFirstInning bool `gorm:"default:false" json:"scored_first_inning"` // Marcó en primer inning (béisbol)
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
