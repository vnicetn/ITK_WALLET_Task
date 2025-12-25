package models

import (
	"time"

	"github.com/google/uuid"
)

type WalletOperation struct {
	WalletID  uuid.UUID     `json:"valletId"`
	Operation OperationType `json:"OperationType"`
	Amount    uint          `json:"amount"`
	CreatedAt time.Time     `json:"createdAt"`
}

type OperationType string

const (
	OperationTypeWithdraw OperationType = "WITHDRAW"
	OperationTypeDeposit  OperationType = "DEPOSIT"
)

type Wallet struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Balance   uint      `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
