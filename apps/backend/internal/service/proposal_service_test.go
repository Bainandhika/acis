package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
)

type mockProposalRepo struct {
	repository.ProposalRepository
	proposal *domain.Proposal
	err      error
}

func (m *mockProposalRepo) Create(ctx context.Context, exec repository.DBExecutor, proposal *domain.Proposal) error {
	return m.err
}

func (m *mockProposalRepo) RejectProposal(ctx context.Context, exec repository.DBExecutor, proposalID string, reviewerID string) error {
	return m.err
}

func TestRejectProposal(t *testing.T) {
	t.Run("Reject proposal success", func(t *testing.T) {
		repo := &mockProposalRepo{err: nil}
		svc := &proposalService{
			proposalRepo: repo,
		}

		err := svc.RejectProposal(context.Background(), "prop-1", "user-1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Reject proposal error", func(t *testing.T) {
		repo := &mockProposalRepo{err: errors.New("not found")}
		svc := &proposalService{
			proposalRepo: repo,
		}

		err := svc.RejectProposal(context.Background(), "prop-1", "user-1")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}
