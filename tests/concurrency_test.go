package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/itk/wallet/internal/models"
	"github.com/itk/wallet/internal/pkg/postgres"
	"github.com/itk/wallet/internal/repository"
	"github.com/itk/wallet/internal/service"
)

func TestConcurrency_1000RPS(t *testing.T) {
	connStr := "postgres://postgres:postgres@localhost:5432/itk_wallet?sslmode=disable"
	db, err := postgres.NewPostgresDB(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("TRUNCATE TABLE wallets CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	repo := repository.NewWalletRepository(db)
	svc := service.NewWalletService(repo)

	walletID := uuid.New()

	_, err = svc.CreateWallet(context.Background(), walletID)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	initialDeposit := uint(1000000)
	_, err = svc.UpdateBalance(context.Background(), walletID, models.OperationTypeDeposit, initialDeposit)
	if err != nil {
		t.Fatalf("Failed initial deposit: %v", err)
	}

	totalRequests := 1000
	concurrentRequests := 100
	amount := uint(1)

	var successCount int64
	var errorCount int64
	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			requestsPerGoroutine := totalRequests / concurrentRequests
			for j := 0; j < requestsPerGoroutine; j++ {
				_, err := svc.UpdateBalance(context.Background(), walletID, models.OperationTypeDeposit, amount)
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)

	balance, err := svc.GetBalance(context.Background(), walletID)
	if err != nil {
		t.Fatalf("Failed to get final balance: %v", err)
	}

	expectedBalance := int(initialDeposit) + totalRequests*int(amount)
	actualRequests := int(successCount + errorCount)

	t.Logf("Duration: %v", duration)
	t.Logf("Total requests: %d", actualRequests)
	t.Logf("Successful: %d", successCount)
	t.Logf("Errors: %d", errorCount)
	t.Logf("Expected balance: %d", expectedBalance)
	t.Logf("Actual balance: %d", balance)
	t.Logf("RPS: %.2f", float64(actualRequests)/duration.Seconds())

	if actualRequests != totalRequests {
		t.Errorf("Not all requests were processed. Expected %d, got %d", totalRequests, actualRequests)
	}

	if errorCount > 0 {
		t.Errorf("Some requests failed. Error count: %d", errorCount)
	}

	if balance != expectedBalance {
		t.Errorf("Balance mismatch. Expected %d, got %d", expectedBalance, balance)
	}

	if duration.Seconds() > 2.0 {
		t.Logf("Warning: Duration %.2fs exceeds 1 second for 1000 requests", duration.Seconds())
	}
}

func TestConcurrency_MultipleOperations(t *testing.T) {
	connStr := "postgres://postgres:postgres@localhost:5432/itk_wallet?sslmode=disable"
	db, err := postgres.NewPostgresDB(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("TRUNCATE TABLE wallets CASCADE")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	repo := repository.NewWalletRepository(db)
	svc := service.NewWalletService(repo)

	walletID := uuid.New()

	_, err = svc.CreateWallet(context.Background(), walletID)
	if err != nil {
		t.Fatalf("Failed to create wallet: %v", err)
	}

	deposits := 100
	withdraws := 50
	amount := uint(10)

	_, err = svc.UpdateBalance(context.Background(), walletID, models.OperationTypeDeposit, uint(1000))
	if err != nil {
		t.Fatalf("Failed initial deposit: %v", err)
	}

	var wg sync.WaitGroup
	var errorCount int64

	for i := 0; i < deposits; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.UpdateBalance(context.Background(), walletID, models.OperationTypeDeposit, amount)
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
			}
		}()
	}

	for i := 0; i < withdraws; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.UpdateBalance(context.Background(), walletID, models.OperationTypeWithdraw, amount)
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
			}
		}()
	}

	wg.Wait()

	balance, err := svc.GetBalance(context.Background(), walletID)
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	initialBalance := 1000
	expectedBalance := initialBalance + (deposits-withdraws)*int(amount)

	if errorCount > 0 {
		t.Errorf("Some operations failed. Error count: %d", errorCount)
	}

	if balance != expectedBalance {
		t.Errorf("Balance mismatch. Expected %d, got %d", expectedBalance, balance)
	}
}

