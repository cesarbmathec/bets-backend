package dtos

// JoinTournamentRequest define las opciones para unirse a un torneo.
type JoinTournamentRequest struct {
	PayWithTokens bool `json:"pay_with_tokens"` // true para pagar con tokens, false para saldo real
}
