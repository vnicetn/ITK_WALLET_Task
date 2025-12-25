package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/itk/wallet/internal/models"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{
		db: db,
	}
}

type WalletInterface interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (int, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, operationType models.OperationType, amount uint) (bool, error)
	CreateWallet(ctx context.Context, walletID uuid.UUID) (bool, error)
}

func (r *WalletRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (int, error) {
	var balance int
	err := r.db.QueryRowContext(ctx, "SELECT balance from wallets WHERE id = $1", walletID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("balance not found")
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, operationType models.OperationType, amount uint) (bool, error) {
	var balance uint
	var newBalance uint

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, err
	}

	err = tx.QueryRowContext(ctx, "SELECT balance from wallets WHERE id = $1 FOR UPDATE", walletID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("wallet not found")
		}
		return false, fmt.Errorf("failed to get balance: %w", err)
	}

	switch operationType {
	case models.OperationTypeDeposit:
		newBalance = balance + amount
	case models.OperationTypeWithdraw:
		if balance < amount {
			tx.Rollback()
			return false, fmt.Errorf("insufficient funds")
		}
		newBalance = balance - amount
	default:
		return false, fmt.Errorf("invalid operation type")
	}

	result, err := tx.ExecContext(ctx, "UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2", newBalance, walletID)
	if err != nil {
		return false, fmt.Errorf("failed to update wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, fmt.Errorf("failed to update wallet")
	}

	return true, tx.Commit()
}

func (r *WalletRepository) CreateWallet(ctx context.Context, walletID uuid.UUID) (bool, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO wallets (id, balance, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (id) DO NOTHING", walletID, 0)
	if err != nil {
		return false, fmt.Errorf("failed to create wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}
