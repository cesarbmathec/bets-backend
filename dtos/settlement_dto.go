package dtos

// SettleEventRequest defines the data needed to settle an event and calculate points.
type SetEventResultRequest struct {
	Results []CompetitorResult `json:"results" binding:"required,min=1,dive"`
}

// CompetitorResult holds the final outcome for a single competitor.
type CompetitorResult struct {
	CompetitorID uint `json:"competitor_id" binding:"required"`
	FinalScore   int  `json:"final_score"` // Para deportes de equipo (goles, puntos)
	Position     int  `json:"position"`    // Para carreras (1ro, 2do, 3ro)
}
