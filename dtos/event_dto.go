package dtos

import "time"

// CreateEventRequest - Crear un evento global (sin asignar a torneo)
type CreateEventRequest struct {
	Name      string  `json:"name" binding:"required" example:"Barcelona vs Real Madrid"`
	Venue     string  `json:"venue" example:"Camp Nou"`
	Line      float64 `json:"line" example:2.5`
	StartTime string  `json:"start_time" binding:"required" example:"2026-02-28T15:00:00Z"`
}

// UpdateEventRequest - Actualizar un evento
type UpdateEventRequest struct {
	Name      string  `json:"name"`
	Venue     string  `json:"venue"`
	Line      float64 `json:"line"`
	StartTime string  `json:"start_time"`
	Status    string  `json:"status"`
}

// AssignEventToTournamentRequest - Asignar un evento a un torneo
type AssignEventToTournamentRequest struct {
	EventID      uint  `json:"event_id" binding:"required"`
	TournamentID uint  `json:"tournament_id" binding:"required"`
	SessionID    *uint `json:"session_id"`
	Order        int   `json:"order"`
}

// SetEventCompetitorsRequest - Establecer competidores de un evento
type SetEventCompetitorsRequest struct {
	Competitors []EventCompetitorInput `json:"competitors" binding:"required"`
}

// EventCompetitorInput - Datos de un competidor en un evento
type EventCompetitorInput struct {
	CompetitorID   *uint  `json:"competitor_id"`
	Name           string `json:"name" binding:"required"`
	AssignedNumber int    `json:"assigned_number"`
	Odds           int    `json:"odds"`
	Runline        int    `json:"runline"`
	SuperRunline   int    `json:"super_runline"`
	IsFavorite     bool   `json:"is_favorite"`
}

// SettleEventRequest - Liquidar un evento
type SettleEventRequest struct {
	Results []SettleResult `json:"results" binding:"required"`
}

// SettleResult - Resultado de un competidor
type SettleResult struct {
	CompetitorID      uint `json:"competitor_id"`
	FinalScore        int  `json:"final_score"`
	Position          int  `json:"position"`
	IsScratched       bool `json:"is_scratched"`
	ScoredFirst       bool `json:"scored_first"`
	ScoredFirstHalf   bool `json:"scored_first_half"`
	ScoredSecondHalf  bool `json:"scored_second_half"`
	ScoredFirstInning bool `json:"scored_first_inning"`
}

// TournamentEventResponse - Evento asignado a un torneo
type TournamentEventResponse struct {
	ID           uint          `json:"id"`
	EventID      uint          `json:"event_id"`
	TournamentID uint          `json:"tournament_id"`
	SessionID    *uint         `json:"session_id,omitempty"`
	Order        int           `json:"order"`
	Event        EventResponse `json:"event"`
}

// EventResponse - Respuesta de evento
type EventResponse struct {
	ID          uint                      `json:"id"`
	Name        string                    `json:"name"`
	Slug        string                    `json:"slug"`
	Venue       string                    `json:"venue"`
	Line        float64                   `json:"line"`
	StartTime   time.Time                 `json:"start_time"`
	Status      string                    `json:"status"`
	ResultNote  string                    `json:"result_note,omitempty"`
	TotalScore  float64                   `json:"total_score"`
	Competitors []EventCompetitorResponse `json:"competitors,omitempty"`
}

// EventCompetitorResponse - Respuesta de competidor en evento
type EventCompetitorResponse struct {
	ID             uint   `json:"id"`
	CompetitorID   *uint  `json:"competitor_id,omitempty"`
	Name           string `json:"name"`
	AssignedNumber int    `json:"assigned_number"`
	Odds           int    `json:"odds"`
	Runline        int    `json:"runline"`
	SuperRunline   int    `json:"super_runline"`
	IsFavorite     bool   `json:"is_favorite"`
	FinalScore     int    `json:"final_score"`
	Position       int    `json:"position"`
	IsScratched    bool   `json:"is_scratched"`
}
