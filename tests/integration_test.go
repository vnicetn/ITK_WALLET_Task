package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/itk/wallet/internal/models"
	"github.com/itk/wallet/internal/pkg/postgres"
	"github.com/itk/wallet/internal/repository"
	"github.com/itk/wallet/internal/service"
)

func setupTestDB(t *testing.T) *sql.DB {
	connStr := "postgres://postgres:postgres@localhost:5432/itk_wallet?sslmode=disable"
	db, err := postgres.NewPostgresDB(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	_, err = db.Exec("TRUNCATE TABLE wallets CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	return db
}

func TestIntegration_WalletFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewWalletRepository(db)
	svc := service.NewWalletService(repo)

	walletID := uuid.New()

	t.Run("create wallet and deposit", func(t *testing.T) {
		created, err := svc.CreateWallet(context.Background(), walletID)
		if err != nil {
			t.Fatalf("Failed to create wallet: %v", err)
		}
		if !created {
			t.Error("Wallet should be created")
		}

		success, err := svc.UpdateBalance(context.Background(), walletID, models.OperationTypeDeposit, 1000)
		if err != nil {
			t.Fatalf("Failed to deposit: %v", err)
		}
		if !success {
			t.Error("Deposit should succeed")
		}

		balance, err := svc.GetBalance(context.Background(), walletID)
		if err != nil {
			t.Fatalf("Failed to get balance: %v", err)
		}
		if balance != 1000 {
			t.Errorf("Expected balance 1000, got %d", balance)
		}
	})

	t.Run("withdraw funds", func(t *testing.T) {
		success, err := svc.UpdateBalance(context.Background(), walletID, models.OperationTypeWithdraw, 300)
		if err != nil {
			t.Fatalf("Failed to withdraw: %v", err)
		}
		if !success {
			t.Error("Withdraw should succeed")
		}

		balance, err := svc.GetBalance(context.Background(), walletID)
		if err != nil {
			t.Fatalf("Failed to get balance: %v", err)
		}
		if balance != 700 {
			t.Errorf("Expected balance 700, got %d", balance)
		}
	})

	t.Run("insufficient funds", func(t *testing.T) {
		success, err := svc.UpdateBalance(context.Background(), walletID, models.OperationTypeWithdraw, 1000)
		if err == nil {
			t.Error("Expected error for insufficient funds")
		}
		if success {
			t.Error("Withdraw should fail")
		}
		if err != nil {
			errStr := err.Error()
			found := false
			for i := 0; i <= len(errStr)-len("insufficient funds"); i++ {
				if i+len("insufficient funds") <= len(errStr) && errStr[i:i+len("insufficient funds")] == "insufficient funds" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Error should contain 'insufficient funds', got: %s", errStr)
			}
		}
	})
}

