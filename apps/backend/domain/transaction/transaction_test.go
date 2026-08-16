package transaction_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bainandhika/acis/apps/backend/domain/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) CreateTransaction(ctx context.Context, exec transaction.DBExecutor, tx *transaction.Transaction) error {
	args := m.Called(ctx, exec, tx)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetTransactionsByFamilyID(ctx context.Context, familyID string) ([]transaction.Transaction, error) {
	args := m.Called(ctx, familyID)
	return args.Get(0).([]transaction.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetTransactionByID(ctx context.Context, txID string) (*transaction.Transaction, error) {
	args := m.Called(ctx, txID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) DeleteTransaction(ctx context.Context, exec transaction.DBExecutor, txID string) error {
	args := m.Called(ctx, exec, txID)
	return args.Error(0)
}

func (m *MockTransactionRepository) CreateProposal(ctx context.Context, prop *transaction.Proposal) error {
	args := m.Called(ctx, prop)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetProposalsByFamilyID(ctx context.Context, familyID string) ([]transaction.Proposal, error) {
	args := m.Called(ctx, familyID)
	return args.Get(0).([]transaction.Proposal), args.Error(1)
}

func (m *MockTransactionRepository) GetProposalForUpdate(ctx context.Context, exec transaction.DBExecutor, proposalID string) (*transaction.Proposal, error) {
	args := m.Called(ctx, exec, proposalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Proposal), args.Error(1)
}

func (m *MockTransactionRepository) GetWalletForUpdate(ctx context.Context, exec transaction.DBExecutor, walletID string) (*transaction.Wallet, error) {
	args := m.Called(ctx, exec, walletID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Wallet), args.Error(1)
}

func (m *MockTransactionRepository) UpdateWalletBalance(ctx context.Context, exec transaction.DBExecutor, walletID string, newBalance float64) error {
	args := m.Called(ctx, exec, walletID, newBalance)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateProposalStatus(ctx context.Context, exec transaction.DBExecutor, proposalID, status, reviewerID string) error {
	args := m.Called(ctx, exec, proposalID, status, reviewerID)
	return args.Error(0)
}

func TestGetTransactionsByFamilyID(t *testing.T) {
	mockRepo := new(MockTransactionRepository)
	now := time.Now()
	desc := "Groceries"
	uid := "user-1"

	txList := []transaction.Transaction{
		{
			ID:          "tx-100",
			WalletID:    "wallet-1",
			Amount:      50000,
			Type:        "expense",
			Description: &desc,
			CreatedBy:   &uid,
			CreatedAt:   now,
		},
	}

	mockRepo.On("GetTransactionsByFamilyID", mock.Anything, "family-1").Return(txList, nil)

	svc := transaction.NewService(mockRepo, nil, nil)
	result, err := svc.GetTransactions(context.Background(), "family-1")

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "tx-100", result[0].ID)
	assert.Equal(t, 50000.0, result[0].Amount)

	mockRepo.AssertExpectations(t)
}
