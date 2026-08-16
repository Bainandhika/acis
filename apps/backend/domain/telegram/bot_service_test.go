package telegram_test

import (
	"context"
	"errors"
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]telegram.WalletBalanceDTO), args.Error(1)
}

func (m *MockFamilyService) GetLowBalanceWallets(ctx context.Context) ([]telegram.LowBalanceWalletDTO, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]telegram.LowBalanceWalletDTO), args.Error(1)
}

func (m *MockFamilyService) FindByTelegramChatID(ctx context.Context, chatID int64) (*telegram.FamilyDTO, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telegram.FamilyDTO), args.Error(1)
}

func (m *MockFamilyService) LinkTelegramChatID(ctx context.Context, inviteCode string, chatID int64) error {
	args := m.Called(ctx, inviteCode, chatID)
	return args.Error(0)
}

func (m *MockFamilyService) GetMembers(ctx context.Context, familyID string) ([]telegram.FamilyMemberDTO, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]telegram.FamilyMemberDTO), args.Error(1)
}

type MockAuthResolver struct {
	mock.Mock
}

func (m *MockAuthResolver) ResolveAuthSession(ctx context.Context, sessionToken string, chatID int64) (string, error) {
	args := m.Called(ctx, sessionToken, chatID)
	return args.String(0), args.Error(1)
}

func (m *MockAuthResolver) GetActiveOTP(ctx context.Context, email, phone string, chatID int64) (string, error) {
	args := m.Called(ctx, email, phone, chatID)
	return args.String(0), args.Error(1)
}

func TestBotService_StartWithAuthToken(t *testing.T) {
	mockTxSvc := new(MockTransactionService)
	mockFamSvc := new(MockFamilyService)
	mockAuth := new(MockAuthResolver)
	tgClient := telegram.NewClient("")
	botSvc := telegram.NewBotService(mockTxSvc, mockFamSvc, mockAuth, tgClient)

	ctx := context.Background()
	mockAuth.On("ResolveAuthSession", ctx, "auth_session_123", int64(1001)).Return("654321", nil)

	update := &telegram.TelegramUpdate{
		Message: &telegram.TelegramMessage{
			MessageID: 1,
			Chat:      telegram.TelegramChat{ID: 1001},
			Text:      "/start auth_session_123",
		},
	}

	reply, err := botSvc.ProcessUpdate(ctx, update)
	assert.NoError(t, err)
	assert.Contains(t, reply, "654321")
	assert.Contains(t, reply, "Kode Masuk ACIS")
	mockAuth.AssertExpectations(t)
}

func TestBotService_LinkCommand(t *testing.T) {
	mockTxSvc := new(MockTransactionService)
	mockFamSvc := new(MockFamilyService)
	mockAuth := new(MockAuthResolver)
	tgClient := telegram.NewClient("")
	botSvc := telegram.NewBotService(mockTxSvc, mockFamSvc, mockAuth, tgClient)

	ctx := context.Background()

	mockFamSvc.On("LinkTelegramChatID", ctx, "ABC123", int64(1001)).Return(nil)

	update := &telegram.TelegramUpdate{
		Message: &telegram.TelegramMessage{
			MessageID: 1,
			Chat:      telegram.TelegramChat{ID: 1001},
			Text:      "/link ABC123",
		},
	}

	reply, err := botSvc.ProcessUpdate(ctx, update)
	assert.NoError(t, err)
	assert.Contains(t, reply, "berhasil dihubungkan")
	mockFamSvc.AssertExpectations(t)
}

func TestBotService_SaldoUnlinked(t *testing.T) {
	mockTxSvc := new(MockTransactionService)
	mockFamSvc := new(MockFamilyService)
	mockAuth := new(MockAuthResolver)
	tgClient := telegram.NewClient("")
	botSvc := telegram.NewBotService(mockTxSvc, mockFamSvc, mockAuth, tgClient)

	ctx := context.Background()

	mockFamSvc.On("FindByTelegramChatID", ctx, int64(9999)).Return(nil, errors.New("not found"))

	update := &telegram.TelegramUpdate{
		Message: &telegram.TelegramMessage{
			MessageID: 1,
			Chat:      telegram.TelegramChat{ID: 9999},
			Text:      "/saldo",
		},
	}

	reply, err := botSvc.ProcessUpdate(ctx, update)
	assert.NoError(t, err)
	assert.Contains(t, reply, "belum terhubung")
	mockFamSvc.AssertExpectations(t)
}

func TestBotService_CatatCommand(t *testing.T) {
	mockTxSvc := new(MockTransactionService)
	mockFamSvc := new(MockFamilyService)
	mockAuth := new(MockAuthResolver)
	tgClient := telegram.NewClient("")
	botSvc := telegram.NewBotService(mockTxSvc, mockFamSvc, mockAuth, tgClient)

	ctx := context.Background()

	mockFamSvc.On("FindByTelegramChatID", ctx, int64(1001)).Return(&telegram.FamilyDTO{
		ID:   "fam-1",
		Name: "Cemara",
	}, nil)

	mockFamSvc.On("GetMembers", ctx, "fam-1").Return([]telegram.FamilyMemberDTO{
		{ID: "m-1", UserID: "u-admin", Role: "admin"},
	}, nil)

	mockFamSvc.On("GetWalletBalances", ctx, "fam-1").Return([]telegram.WalletBalanceDTO{
		{WalletID: "wallet-1", WalletName: "Makan", CurrentBalance: 100000},
	}, nil)

	desc := "Nasi Goreng"
	mockTxSvc.On("CreateDirectTransaction", ctx, telegram.CreateTransactionDTO{
		WalletID:    "wallet-1",
		UserID:      "u-admin",
		Type:        "expense",
		Amount:      50000,
		Category:    "telegram_catat",
		Description: &desc,
	}).Return(&telegram.TransactionDTO{
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

	mockFamSvc.AssertExpectations(t)
	mockTxSvc.AssertExpectations(t)
}
