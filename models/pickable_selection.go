package models

// Tipos de selección disponibles en el sistema
const (
	SelectionTypeMacho           = "macho"               // Gana el favorito
	SelectionTypeHembra          = "hembra"              // Gana el no favorito
	SelectionTypeMachoRL         = "macho_rl"            // Gana el favorito con runline
	SelectionTypeHembraRL        = "hembra_rl"           // Gana el no favorito con runline
	SelectionTypeMachoSRL        = "macho_srl"           // Gana el favorito con super runline
	SelectionTypeHembraSRL       = "hembra_srl"          // Gana el no favorito con super runline
	SelectionTypeAlta            = "alta"                // Over - suma mayor a la línea
	SelectionTypeBaja            = "baja"                // Under - suma menor a la línea
	SelectionTypeEmpate          = "empate"              // Empate
	SelectionTypeSuperAlta       = "super_alta"          // Super over - suma mayor a super línea
	SelectionTypeSuperBaja       = "super_baja"          // Super under - suma menor a super línea
	SelectionTypeMarcaPrimero    = "marca_primero"       // Equipo que marca primero
	SelectionTypeMarcaPrimeroT   = "marca_primer_tiempo" // Equipo que marca en primer tiempo/cuarto
	SelectionTypePrimeraMitad    = "primera_mitad"       // Gana en primera mitad
	SelectionTypeSegundaMitad    = "segunda_mitad"       // Gana en segunda mitad
	SelectionTypeCarreraPosicion = "carrera_posicion"    // Posición en carrera de caballos
)

// PickableSelection define una opción de pronóstico configurable por el admin para un evento.
// Ej: "Gana Real Madrid", "Alta de 2.5 goles", "Runline -1.5".
type PickableSelection struct {
	BaseModel
	EventID     uint   `gorm:"index;not null" json:"event_id"`
	Event       Event  `gorm:"foreignKey:EventID" json:"-"`
	Description string `gorm:"size:255;not null" json:"description"` // "Gana Real Madrid" o "Alta (Over) 2.5"

	// Tipo de selección para categorizar
	// Opciones: macho, hembra, macho_rl, hembra_rl, macho_srl, hembra_srl, alta, baja,
	//           empate, super_alta, super_baja, marca_primero, marca_primer_tiempo,
	//           primera_mitad, segunda_mitad, carrera_posicion
	SelectionType string `gorm:"size:50;index" json:"selection_type"`

	// Para selecciones de tipo Alta/Baja (Over/Under)
	// Line: el score total que divide alta de baja (ej: 2.5)
	Line float64 `gorm:"type:decimal(10,2)" json:"line,omitempty"`

	// Para Runline y Superrunline (puntos a favor/en contra)
	// Home: para el equipo local, Away: para el visitante
	RunlineHome    float64 `gorm:"type:decimal(10,2)" json:"runline_home,omitempty"` // ej: -1.5, -2.5
	RunlineAway    float64 `gorm:"type:decimal(10,2)" json:"runline_away,omitempty"` // ej: +1.5, +2.5
	IsSuperRunline bool    `gorm:"default:false" json:"is_super_runline"`            // true si es superrunline

	// Para Macho/Hembra (favorito/no favorito) - Odds del favorito
	// Positive odds: +120, +200, +300 (paga más)
	// Negative odds: -120, -400 (necesita stake mayor)
	Odds int `gorm:"default:0" json:"odds"` // Ej: 120, -400, etc.

	// Para selecciones de tipo posición en carreras de caballos
	// position_for_points: posición que otorga puntos (1=primero, 2=segundo, 3=tercero)
	PositionForPoints int `gorm:"default:1" json:"position_for_points"`

	// Para Macho/Hembra, se asocia al competidor que representa
	CompetitorID *uint `gorm:"index" json:"competitor_id,omitempty"`

	// Puntos a otorgar por acertar esta selección
	PointsForWin  int `gorm:"not null" json:"points_for_win"`
	PointsForPush int `gorm:"default:0" json:"points_for_push"` // Para empates contra la línea

	// Para carreras de caballos - puntos por posición
	PointsForFirst  int `gorm:"default:0" json:"points_for_first"`  // Puntos si queda 1ro
	PointsForSecond int `gorm:"default:0" json:"points_for_second"` // Puntos si queda 2do
	PointsForThird  int `gorm:"default:0" json:"points_for_third"`  // Puntos si queda 3ro

	// Estado final de esta selección después de liquidar el evento.
	Status string `gorm:"size:20;default:'pending'" json:"status"` // pending, won, lost, push, cancelled
}

// TableName define el nombre de la tabla en la base de datos.
func (PickableSelection) TableName() string {
	return "pickable_selections"
}
