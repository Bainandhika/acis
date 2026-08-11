package telegram_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/Bainandhika/acis/apps/backend/internal/family"
	"github.com/Bainandhika/acis/apps/backend/internal/telegram"
	"github.com/Bainandhika/acis/apps/backend/internal/transaction"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) CreateDirectTransaction(ctx context.Context, req transaction.CreateTransactionDTO) (*transaction.TransactionDTO, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.TransactionDTO), args.Error(1)
}

func (m *MockTransactionService) GetTransactions(ctx context.Context, familyID string) ([]transaction.TransactionDTO, error) {
	args := m.Called(ctx, familyID)
	return args.Get(0).([]transaction.TransactionDTO), args.Error(1)
}

func (m *MockTransactionService) CreateProposal(ctx context.Context, req transaction.CreateProposalDTO) (*transaction.ProposalDTO, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*transaction.ProposalDTO), args.Error(1)
}

func (m *MockTransactionService) ApproveProposal(ctx context.Context, proposalID string, reviewerID string) error {
	args := m.Called(ctx, proposalID, reviewerID)
	return args.Error(0)
}

func (m *MockTransactionService) RejectProposal(ctx context.Context, proposalID string, reviewerID string) error {
	args := m.Called(ctx, proposalID, reviewerID)
	return args.Error(0)
}

type MockFamilyService struct {
	mock.Mock
}

func (m *MockFamilyService) CreateFamily(ctx context.Context, userID, name string) (*family.FamilyDTO, error) {
	args := m.Called(ctx, userID, name)
	return args.Get(0).(*family.FamilyDTO), args.Error(1)
}

func (m *MockFamilyService) JoinFamily(ctx context.Context, userID, inviteCode string) (*family.FamilyDTO, error) {
	args := m.Called(ctx, userID, inviteCode)
	return args.Get(0).(*family.FamilyDTO), args.Error(1)
}

func (m *MockFamilyService) GetMyFamily(ctx context.Context, userID string) (*family.FamilyDTO, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*family.FamilyDTO), args.Error(1)
}

func (m *MockFamilyService) CreateWallet(ctx context.Context, userID, familyID string, req family.CreateWalletReq) (*family.WalletDTO, error) {
	args := m.Called(ctx, userID, familyID, req)
	return args.Get(0).(*family.WalletDTO), args.Error(1)
}

func (m *MockFamilyService) GetWallets(ctx context.Context, familyID string) ([]family.WalletDTO, error) {
	args := m.Called(ctx, familyID)
	return args.Get(0).([]family.WalletDTO), args.Error(1)
}

func (m *MockFamilyService) GetWalletBalances(ctx context.Context, familyID string) ([]family.WalletBalanceDTO, error) {
	args := m.Called(ctx, familyID)
	return args.Get(0).([]family.WalletBalanceDTO), args.Error(1)
}

func (m *MockFamilyService) GetLowBalanceWallets(ctx context.Context) ([]family.LowBalanceWalletDTO, error) {
	args := m.Called(ctx)
	return args.Get(0).([]family.LowBalanceWalletDTO), args.Error(1)
}

func TestBotService_CatatCommand(t *testing.T) {
	mockTxSvc := new(MockTransactionService)
	mockFamSvc := new(MockFamilyService)
	tgClient := telegram.NewClient("")
	botSvc := telegram.NewBotService(mockTxSvc, mockFamSvc, tgClient)

	ctx := context.Background()

	mockTxSvc.On("CreateDirectTransaction", ctx, mock.Anything).Return(&transaction.TransactionDTO{
		ID:     "tx-123",
		Amount: 50000,
	}, nil)

	update := &telegram.TelegramUpdate{
		Message: &telegram.TelegramMessage{
			MessageID: 1,
			Chat:      telegram.TelegramChat{ID: 1001},
			Text:      "/catat wallet-1 50000 Nasi Goreng",
		},
	}

	reply, err := botSvc.ProcessUpdate(ctx, update)
	assert.NoError(t, err)
	assert.Contains(t, reply, "Transaksi tersimpan")
	assert.Contains(t, reply, "tx-123")

	mockTxSvc.AssertExpectations(t)
}
