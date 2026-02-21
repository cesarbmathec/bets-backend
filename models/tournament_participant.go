package models

// TournamentParticipant representa la inscripción de un usuario a un torneo y su puntaje.
type TournamentParticipant struct {
	BaseModel
	UserID       uint `gorm:"uniqueIndex:idx_user_tournament;not null" json:"user_id"`
	TournamentID uint `gorm:"uniqueIndex:idx_user_tournament;not null" json:"tournament_id"`
	TotalPoints  int  `gorm:"default:0" json:"total_points"`

	User       User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tournament Tournament `gorm:"foreignKey:TournamentID" json:"-"`
	UserPicks  []UserPick `gorm:"foreignKey:ParticipantID" json:"picks,omitempty"`
}

// TableName define el nombre de la tabla en la base de datos.
func (TournamentParticipant) TableName() string {
	return "tournament_participants"
}
