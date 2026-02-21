package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Session representa una jornada/día de apuestas dentro de un torneo.
// Ejemplo: Sesión 1 = Lunes, Sesión 2 = Martes, etc.
// Cada sesión tiene un rango de tiempo donde los participantes pueden hacer sus selecciones.
type Session struct {
	BaseModel
	TournamentID  uint       `gorm:"index;not null" json:"tournament_id"`
	Tournament    Tournament `gorm:"foreignKey:TournamentID" json:"-"`
	SessionNumber int        `gorm:"not null" json:"session_number"` // 1, 2, 3, 4, 5...

	// Fecha y hora de inicio de la sesión (cuando abre para hacer selecciones)
	StartTime time.Time `gorm:"not null" json:"start_time"`

	// Fecha y hora límite de la sesión (hora tope para hacer selecciones, antes de que inicien los partidos)
	EndTime time.Time `gorm:"not null" json:"end_time"`

	// Descripción opcional de la sesión (ej: "Jornada de Lunes - 5 partidos")
	Description string `gorm:"size:255" json:"description"`

	// Estado de la sesión: open (abierta para picks), closed (cerrada), settled (liquidada)
	Status string `gorm:"size:20;default:'open';index" json:"status"`

	// Relaciones
	Events []Event `gorm:"foreignKey:SessionID" json:"events,omitempty"`
}

// Hook BeforeSave para validaciones
func (s *Session) BeforeSave(tx *gorm.DB) (err error) {
	// Validar que EndTime sea posterior a StartTime
	if s.EndTime.Before(s.StartTime) {
		return errors.New("la hora de cierre no puede ser anterior a la de inicio")
	}

	// Validar que el número de sesión sea positivo
	if s.SessionNumber < 1 {
		return errors.New("el número de sesión debe ser mayor a 0")
	}

	return nil
}

// TableName define el nombre de la tabla
func (Session) TableName() string {
	return "tournament_sessions"
}
