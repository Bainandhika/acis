package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
)

type mockTxRepo struct {
	repository.TransactionRepository
	list []domain.Transaction
	err  error
}

func (m *mockTxRepo) GetByWalletID(ctx context.Context, exec repository.DBExecutor, walletID string, limit, offset int) ([]domain.Transaction, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.list, nil
}

func TestGetTransactions(t *testing.T) {
	now := time.Now()
	desc := "Nasi Goreng"
	creator := "user-1"

	repo := &mockTxRepo{
		list: []domain.Transaction{
			{
				ID:          "tx-1",
				WalletID:    "w-1",
				Amount:      15000,
				Type:        "expense",
				Description: &desc,
				CreatedBy:   &creator,
				CreatedAt:   now,
			},
		},
	}

	svc := &transactionService{
		txRepo: repo,
	}

	res, err := svc.GetTransactions(context.Background(), "w-1", 10, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(res))
	}

	if res[0].ID != "tx-1" || res[0].Amount != 15000 {
		t.Errorf("Unexpected result data: %+v", res[0])
	}
}

func TestGetTransactionsError(t *testing.T) {
	repo := &mockTxRepo{err: errors.New("db query error")}
	svc := &transactionService{txRepo: repo}

	_, err := svc.GetTransactions(context.Background(), "w-1", 10, 0)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
