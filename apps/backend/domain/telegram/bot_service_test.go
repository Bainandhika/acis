package telegram_test

import (
	"context"
	"testing"

	"github.com/Bainandhika/acis/apps/backend/domain/telegram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) CreateDirectTransaction(ctx context.Context, req telegram.CreateTransactionDTO) (*telegram.TransactionDTO, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telegram.TransactionDTO), args.Error(1)
}

type MockFamilyService struct {
	mock.Mock
}

func (m *MockFamilyService) GetWalletBalances(ctx context.Context, familyID string) ([]telegram.WalletBalanceDTO, error) {
	args := m.Called(ctx, familyID)
	return args.Get(0).([]telegram.WalletBalanceDTO), args.Error(1)
}

func (m *MockFamilyService) GetLowBalanceWallets(ctx context.Context) ([]telegram.LowBalanceWalletDTO, error) {
	args := m.Called(ctx)
	return args.Get(0).([]telegram.LowBalanceWalletDTO), args.Error(1)
}

func TestBotService_CatatCommand(t *testing.T) {
	mockTxSvc := new(MockTransactionService)
	mockFamSvc := new(MockFamilyService)
	tgClient := telegram.NewClient("")
	botSvc := telegram.NewBotService(mockTxSvc, mockFamSvc, tgClient)

	ctx := context.Background()

	mockTxSvc.On("CreateDirectTransaction", ctx, mock.Anything).Return(&telegram.TransactionDTO{
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
