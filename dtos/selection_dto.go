package dtos

type CreateSelectionRequest struct {
	EventID       uint    `json:"event_id" binding:"required"`
	Description   string  `json:"description" binding:"required"`
	SelectionType string  `json:"selection_type" binding:"required"` // "Ganador", "Alta", "Baja", "Macho", "Hembra", "Runline", "Superrunline", "Empate"
	Line          float64 `json:"line"`                              // Ej: 2.5 (para altas/bajas)

	// Para Runline y Superrunline
	RunlineHome    float64 `json:"runline_home"`
	RunlineAway    float64 `json:"runline_away"`
	IsSuperRunline bool    `json:"is_super_runline"`

	// Para Macho/Hembra - Odds del favorito
	// Positive: +120, +200, +300
	// Negative: -120, -400
	Odds int `json:"odds"`

	CompetitorID  *uint `json:"competitor_id"` // Opcional, si es apuesta a ganador
	PointsForWin  int   `json:"points_for_win" binding:"required"`
	PointsForPush int   `json:"points_for_push"`
}

type SubmitPicksRequest struct {
	SelectionIDs []uint `json:"selection_ids" binding:"required,min=1"`
}

/*
type SubmitPicksBySessionRequest struct {
	SessionID    uint   `json:"session_id" binding:"required"`
	SelectionIDs []uint `json:"selection_ids" binding:"required,min=1"`
}
*/
