package dtos

import "time"

// CreateTournamentRequest define los datos necesarios para crear un nuevo torneo.
type CreateTournamentRequest struct {
	Name            string                    `json:"name" binding:"required"`
	Description     string                    `json:"description"`
	Category        string                    `json:"category" binding:"required"` // "Hipica", "Futbol", etc.
	StartDate       time.Time                 `json:"start_date" binding:"required"`
	EndDate         time.Time                 `json:"end_date" binding:"required"`
	EntryFee        float64                   `json:"entry_fee" binding:"gte=0"`
	EntryFeeTokens  int                       `json:"entry_fee_tokens" binding:"gte=0"`
	PrizeBonus      float64                   `json:"prize_bonus" binding:"gte=0"`
	AdminFeePercent float64                   `json:"admin_fee_percent" binding:"gte=0,lte=100"`
	Settings        TournamentSettingsRequest `json:"settings"`
}

// TournamentSettingsRequest define las reglas específicas del torneo en la creación.
type TournamentSettingsRequest struct {
	PrizeDistribution      []float64 `json:"prize_distribution"`       // Ej: [0.7, 0.2, 0.1]
	SelectionsPerSession   int       `json:"selections_per_session"`   // Ej: 5 (macho, hembra, alta, baja, runline)
	HorseRacingPoints      []int     `json:"horse_racing_points"`      // Ej: [10, 5, 3]
	RequiredSelectionTypes []string  `json:"required_selection_types"` // Ej: ["macho", "hembra", "alta", "baja", "runline"]
	TotalSessions          int       `json:"total_sessions"`           // Ej: 5 (Lunes a Viernes)
}

// UpdateStatusRequest define el cuerpo para actualizar el estado de un torneo.
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open closed finished"`
}
