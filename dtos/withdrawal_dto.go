package dtos

import "time"

// Request DTOs

// WithdrawalRequest represents a withdrawal request
type WithdrawalRequest struct {
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethodID uint    `json:"payment_method_id" binding:"required"`
}

// VerifyWithdrawalRequest represents verification of a withdrawal with code
type VerifyWithdrawalRequest struct {
	WithdrawalID uint   `json:"withdrawal_id" binding:"required"`
	Code         string `json:"code" binding:"required,len=6"`
}

// CancelWithdrawalRequest represents cancellation of a pending withdrawal
type CancelWithdrawalRequest struct {
	WithdrawalID uint `json:"withdrawal_id" binding:"required"`
}

// Response DTOs

// WithdrawalResponse represents a withdrawal in responses
type WithdrawalResponse struct {
	ID              uint                      `json:"id"`
	Amount          float64                   `json:"amount"`
	PreviousBalance float64                   `json:"previous_balance"`
	NewBalance      float64                   `json:"new_balance"`
	Status          string                    `json:"status"`
	Verified        bool                      `json:"verified"`
	VerifiedAt      *time.Time                `json:"verified_at,omitempty"`
	RejectedReason  string                    `json:"rejected_reason,omitempty"`
	ProcessedAt     *time.Time                `json:"processed_at,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	PaymentMethod   UserPaymentMethodResponse `json:"payment_method"`
}

// WithdrawalWithCodeResponse includes the verification code (only shown once)
type WithdrawalWithCodeResponse struct {
	ID             uint    `json:"id"`
	Amount         float64 `json:"amount"`
	Status         string  `json:"status"`
	WithdrawalCode string  `json:"withdrawal_code"`
	Message        string  `json:"message"`
	ExpiresIn      int     `json:"expires_in_minutes"` // Minutes until code expires
}

// WithdrawalHistoryResponse represents the withdrawal history
type WithdrawalHistoryResponse struct {
	Withdrawals []WithdrawalResponse `json:"withdrawals"`
	Total       int                  `json:"total"`
	Pending     int                  `json:"pending"`
	Approved    int                  `json:"approved"`
	Rejected    int                  `json:"rejected"`
	Completed   int                  `json:"completed"`
}

// VerifyCodeResponse represents verification result
type VerifyCodeResponse struct {
	Valid    bool   `json:"valid"`
	Message  string `json:"message"`
	Attempts int    `json:"attempts_remaining"`
}

// WithdrawalLimitResponse represents user's withdrawal limits
type WithdrawalLimitResponse struct {
	MaxWithdrawalPerDay   float64 `json:"max_withdrawal_per_day"`
	MaxWithdrawalPerWeek  float64 `json:"max_withdrawal_per_week"`
	MaxWithdrawalPerMonth float64 `json:"max_withdrawal_per_month"`
	UsedToday             float64 `json:"used_today"`
	UsedThisWeek          float64 `json:"used_this_week"`
	UsedThisMonth         float64 `json:"used_this_month"`
	AvailableToday        float64 `json:"available_today"`
	AvailableThisWeek     float64 `json:"available_this_week"`
	AvailableThisMonth    float64 `json:"available_this_month"`
}
