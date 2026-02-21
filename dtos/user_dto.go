package dtos

// UpdateUserRoleRequest define los datos para actualizar el rol de un usuario.
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

// UpdateUserStatusRequest define los datos para activar/desactivar un usuario.
type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}
