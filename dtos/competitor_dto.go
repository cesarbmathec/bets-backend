package dtos

// CreateCompetitorRequest define los datos para crear un competidor global.
type CreateCompetitorRequest struct {
	Name           string `json:"name" binding:"required" example:"Real Madrid"`
	Category       string `json:"category" example:"Fútbol"`
	AssignedNumber int    `json:"assigned_number" example:"1"`
	Description    string `json:"description"`
}

// UpdateCompetitorRequest define los datos para actualizar un competidor.
type UpdateCompetitorRequest struct {
	Name           string `json:"name"`
	Category       string `json:"category"`
	AssignedNumber int    `json:"assigned_number"`
	Description    string `json:"description"`
	Status         string `json:"status" binding:"oneof=active inactive"`
}
