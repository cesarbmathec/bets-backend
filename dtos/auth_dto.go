package dtos

// LoginRequest estructura para el inicio de sesión
// Acepta email O username para autenticación
type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse estructura de respuesta tras un login exitoso
type LoginResponse struct {
	Token string      `json:"token"`
	User  UserSummary `json:"user"`
}

// RegisterRequest estructura para el registro de nuevos apostadores
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	// Datos personales
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
	DocumentID string `json:"document_id"`
}

// UserSummary información básica del usuario para el frontend
type UserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
