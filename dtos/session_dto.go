package dtos

import "time"

// CreateSessionRequest define los datos necesarios para crear una sesión dentro de un torneo.
type CreateSessionRequest struct {
	TournamentID  uint      `json:"tournament_id" binding:"required"`
	SessionNumber int       `json:"session_number" binding:"required,min=1"` // 1, 2, 3...
	StartTime     time.Time `json:"start_time" binding:"required"`           // Cuándo abre la sesión
	EndTime       time.Time `json:"end_time" binding:"required"`             // Hora límite (antes de partidos)
	Description   string    `json:"description"`                             // Opcional
}

// UpdateSessionStatusRequest define el cuerpo para actualizar el estado de una sesión.
type UpdateSessionStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open closed settled"`
}

// SubmitPicksBySessionRequest define los datos para enviar predicciones de una sesión específica.
type SubmitPicksBySessionRequest struct {
	SessionID    uint   `json:"session_id" binding:"required"` // ID de la sesión
	SelectionIDs []uint `json:"selection_ids" binding:"required,min=1"`
}

// SessionResponse define la estructura de respuesta para una sesión.
type SessionResponse struct {
	ID            uint      `json:"id"`
	TournamentID  uint      `json:"tournament_id"`
	SessionNumber int       `json:"session_number"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	EventCount    int       `json:"event_count"` // Cantidad de eventos en esta sesión
	PickCount     int       `json:"pick_count"`  // Cantidad de predicciones realizadas por el usuario
}
