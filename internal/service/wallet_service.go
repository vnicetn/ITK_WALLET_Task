package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/itk/wallet/internal/models"
	"github.com/itk/wallet/internal/repository"
)

type WalletService struct {
	walletRepo repository.WalletInterface
}

func NewWalletService(walletRepo repository.WalletInterface) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
	}
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int, error) {
	balance, err := s.walletRepo.GetBalance(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("failed to get wallets balance: %w", err)
	}

	return balance, nil
}

func (s *WalletService) UpdateBalance(ctx context.Context, walletID uuid.UUID, operationType models.OperationType, amount int) (bool, error) {
	if operationType != models.OperationTypeDeposit && operationType != models.OperationTypeWithdraw {
		return false, fmt.Errorf("invalid operationType: %s", operationType)
	}

	if amount <= 0 {
		return false, fmt.Errorf("amount must be greater than zero")
	}

	ok, err := s.walletRepo.UpdateBalance(ctx, walletID, operationType, amount)
	if err != nil {
		return false, fmt.Errorf("failed to update balance: %w", err)
	}

	return ok, nil
}

func (s *WalletService) CreateWallet(ctx context.Context, walletID uuid.UUID) (bool, error) {
	ok, err := s.walletRepo.CreateWallet(ctx, walletID)
	if err != nil {
		return false, fmt.Errorf("failed to create wallet: %w", err)
	}

	return ok, nil
}
