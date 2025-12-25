package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/itk/wallet/internal/models"
)

type MockWalletRepository struct {
	GetBalanceFunc    func(ctx context.Context, walletID uuid.UUID) (int, error)
	UpdateBalanceFunc func(ctx context.Context, walletID uuid.UUID, operationType models.OperationType, amount uint) (bool, error)
	CreateWalletFunc  func(ctx context.Context, walletID uuid.UUID) (bool, error)
}

func (m *MockWalletRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (int, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, walletID)
	}
	return 0, nil
}

func (m *MockWalletRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, operationType models.OperationType, amount uint) (bool, error) {
	if m.UpdateBalanceFunc != nil {
		return m.UpdateBalanceFunc(ctx, walletID, operationType, amount)
	}
	return true, nil
}

func (m *MockWalletRepository) CreateWallet(ctx context.Context, walletID uuid.UUID) (bool, error) {
	if m.CreateWalletFunc != nil {
		return m.CreateWalletFunc(ctx, walletID)
	}
	return true, nil
}

func TestWalletService_GetBalance(t *testing.T) {
	tests := []struct {
		name      string
		walletID  uuid.UUID
		mockSetup func(*MockWalletRepository)
		want      int
		wantErr   bool
	}{
		{
			name:     "successful get balance",
			walletID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			mockSetup: func(m *MockWalletRepository) {
				m.GetBalanceFunc = func(ctx context.Context, walletID uuid.UUID) (int, error) {
					return 1000, nil
				}
			},
			want:    1000,
			wantErr: false,
		},
		{
			name:     "wallet not found",
			walletID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			mockSetup: func(m *MockWalletRepository) {
				m.GetBalanceFunc = func(ctx context.Context, walletID uuid.UUID) (int, error) {
					return 0, errors.New("balance not found")
				}
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockWalletRepository{}
			tt.mockSetup(mockRepo)

			service := &WalletService{walletRepo: mockRepo}
			balance, err := service.GetBalance(context.Background(), tt.walletID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if balance != tt.want {
					t.Errorf("got balance %d, want %d", balance, tt.want)
				}
			}
		})
	}
}

func TestWalletService_UpdateBalance(t *testing.T) {
	walletID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name          string
		operationType models.OperationType
		amount        uint
		mockSetup     func(*MockWalletRepository)
		wantErr       bool
		errContains   string
	}{
		{
			name:          "successful deposit",
			operationType: models.OperationTypeDeposit,
			amount:        1000,
			mockSetup: func(m *MockWalletRepository) {
				m.UpdateBalanceFunc = func(ctx context.Context, walletID uuid.UUID, operationType models.OperationType, amount uint) (bool, error) {
					return true, nil
				}
			},
			wantErr: false,
		},
		{
			name:          "successful withdraw",
			operationType: models.OperationTypeWithdraw,
			amount:        500,
			mockSetup: func(m *MockWalletRepository) {
				m.UpdateBalanceFunc = func(ctx context.Context, walletID uuid.UUID, operationType models.OperationType, amount uint) (bool, error) {
					return true, nil
				}
			},
			wantErr: false,
		},
		{
			name:          "insufficient funds",
			operationType: models.OperationTypeWithdraw,
			amount:        2000,
			mockSetup: func(m *MockWalletRepository) {
				m.UpdateBalanceFunc = func(ctx context.Context, walletID uuid.UUID, operationType models.OperationType, amount uint) (bool, error) {
					return false, errors.New("insufficient funds")
				}
			},
			wantErr:     true,
			errContains: "insufficient funds",
		},
		{
			name:          "invalid operation type",
			operationType: models.OperationType("INVALID"),
			amount:        1000,
			mockSetup:     func(m *MockWalletRepository) {},
			wantErr:       true,
			errContains:   "invalid operationType",
		},
		{
			name:          "zero amount",
			operationType: models.OperationTypeDeposit,
			amount:        0,
			mockSetup:     func(m *MockWalletRepository) {},
			wantErr:       true,
			errContains:   "amount cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockWalletRepository{}
			tt.mockSetup(mockRepo)

			service := &WalletService{walletRepo: mockRepo}
			_, err := service.UpdateBalance(context.Background(), walletID, tt.operationType, tt.amount)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.errContains != "" && err != nil {
					errStr := err.Error()
					found := false
					for i := 0; i <= len(errStr)-len(tt.errContains); i++ {
						if i+len(tt.errContains) <= len(errStr) && errStr[i:i+len(tt.errContains)] == tt.errContains {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("error message should contain '%s', got '%s'", tt.errContains, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
