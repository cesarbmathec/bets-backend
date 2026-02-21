package dtos

import "time"

type WalletResponse struct {
	Balance        float64 `json:"balance" example:"150.50"`
	Bonus          float64 `json:"bonus" example:"10.00"`
	Frozen         float64 `json:"frozen" example:"25.00"`
	TotalAvailable float64 `json:"total_available" example:"160.50"`
	Currency       string  `json:"currency" example:"USD"`
}

type TransactionResponse struct {
	TransactionNumber string    `json:"transaction_number" example:"TRX-20240520-1"`
	Amount            float64   `json:"amount" example:"50.00"`
	Type              string    `json:"type" example:"deposit"` // deposit, bet_payment, prize
	Status            string    `json:"status" example:"completed"`
	Description       string    `json:"description" example:"Depósito inicial"`
	CreatedAt         time.Time `json:"created_at"`
	NewBalance        float64   `json:"new_balance" example:"150.50"`
}

type UserStatsResponse struct {
	TotalDepositsCount    int64   `json:"total_deposits_count"`
	TotalWithdrawalsCount int64   `json:"total_withdrawals_count"`
	TotalWinnings         float64 `json:"total_winnings"`      // Total ganado en premios
	TotalSpent            float64 `json:"total_spent_entries"` // Total gastado en inscripciones
}
