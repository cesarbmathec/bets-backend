package models

// PickableSelection define una opción de pronóstico configurable por el admin para un evento.
// Ej: "Gana Real Madrid", "Alta de 2.5 goles", "Runline -1.5".
type PickableSelection struct {
	BaseModel
	EventID     uint   `gorm:"index;not null" json:"event_id"`
	Event       Event  `gorm:"foreignKey:EventID" json:"-"`
	Description string `gorm:"size:255;not null" json:"description"` // "Gana Real Madrid" o "Alta (Over) 2.5"

	// Tipo de selección para categorizar (Macho, Hembra, Alta, Baja, Runline, Superrunline, Empate)
	SelectionType string `gorm:"size:50;index" json:"selection_type"`

	// Para selecciones de tipo Alta/Baja (Over/Under)
	// Line: el score total que divide alta de baja (ej: 2.5)
	Line float64 `gorm:"type:decimal(10,2)" json:"line,omitempty"`

	// Para Runline y Superrunline (puntos a favor/en contra)
	RunlineHome    float64 `gorm:"type:decimal(10,2)" json:"runline_home,omitempty"` // ej: -1.5, -2.5
	RunlineAway    float64 `gorm:"type:decimal(10,2)" json:"runline_away,omitempty"` // ej: +1.5, +2.5
	IsSuperRunline bool    `gorm:"default:false" json:"is_super_runline"`            // true si es superrunline

	// Para Macho/Hembra (favorito/no favorito) - Odds del favorito
	// Positive odds: +120, +200, +300 (paga más)
	// Negative odds: -120, -400 (necesita stake mayor)
	Odds int `gorm:"default:0" json:"odds"` // Ej: 120, -400, etc.

	// Para selecciones de ganador (Macho/Hembra), se asocia al competidor.
	CompetitorID *uint `gorm:"index" json:"competitor_id,omitempty"`

	// Puntos a otorgar
	PointsForWin  int `gorm:"not null" json:"points_for_win"`
	PointsForPush int `gorm:"default:0" json:"points_for_push"` // Para empates contra la línea

	// Estado final de esta selección después de liquidar el evento.
	Status string `gorm:"size:20;default:'pending'" json:"status"` // pending, won, lost, push
}

// TableName define el nombre de la tabla en la base de datos.
func (PickableSelection) TableName() string {
	return "pickable_selections"
}
