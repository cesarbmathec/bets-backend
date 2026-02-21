package dtos

import "time"

type CreateEventRequest struct {
	TournamentID uint      `json:"tournament_id" binding:"required"`
	Name         string    `json:"name" binding:"required" example:"Carrera de la Amistad"`
	Order        int       `json:"order" example:"1"`
	StartTime    time.Time `json:"start_time" binding:"required"`
}

type SetCompetitorsRequest struct {
	Competitors []CompetitorDetail `json:"competitors" binding:"required,min=2"`
}

type CompetitorDetail struct {
	Name           string `json:"name" binding:"required"`
	AssignedNumber int    `json:"assigned_number"`
}
