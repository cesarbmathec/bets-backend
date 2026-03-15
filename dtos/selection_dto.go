package dtos

type CreateSelectionRequest struct {
	EventID       uint    `json:"event_id" binding:"required"`
	Description   string  `json:"description" binding:"required"`
	SelectionType string  `json:"selection_type" binding:"required"` // "macho", "hembra", "alta", "baja", "macho_runline", "hembra_runline", "macho_srl", "hembra_srl"
	Line          float64 `json:"line"`                              // Ej: 2.5 (para altas/bajas/runlines/superlines)

	// Para Runline y Superrunline
	RunlineHome    float64 `json:"runline_home"`
	RunlineAway    float64 `json:"runline_away"`
	IsSuperRunline bool    `json:"is_super_runline"`

	// Para Macho/Hembra - Odds del favorito
	// Positive: +120, +200, +300
	// Negative: -120, -400
	Odds int `json:"odds"`

	CompetitorID  *uint `json:"competitor_id"`  // Opcional, si es apuesta a ganador
	PointsForWin  int   `json:"points_for_win"` // Opcional: se hereda del Tournament si no se especifica
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
