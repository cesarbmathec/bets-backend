package models

// UserPick es la elección que un participante hace para una selección disponible.
type UserPick struct {
	BaseModel
	ParticipantID uint `gorm:"index;not null" json:"participant_id"` // ID de la tabla TournamentParticipant
	SelectionID   uint `gorm:"index;not null" json:"selection_id"`   // ID de la PickableSelection
	SessionID     uint `gorm:"index;not null" json:"session_id"`     // ID de la sesión a la que pertenece esta selección

	// Estado final de la selección del usuario
	Status        string `gorm:"size:20;default:'pending'" json:"status"` // pending, won, lost, push
	AwardedPoints int    `gorm:"default:0" json:"awarded_points"`

	Participant TournamentParticipant `gorm:"foreignKey:ParticipantID" json:"-"`
	Selection   PickableSelection     `gorm:"foreignKey:SelectionID" json:"selection"`
	Session     Session               `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

func (UserPick) TableName() string {
	return "user_picks"
}
