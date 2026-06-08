package handler

import (
	"net/http"

	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ProposalHandler struct {
	proposalService service.ProposalService
}

func NewProposalHandler(proposalService service.ProposalService) *ProposalHandler {
	return &ProposalHandler{proposalService: proposalService}
}

func (h *ProposalHandler) CreateProposal(c *gin.Context) {
	var req dto.CreateProposalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	proposal, err := h.proposalService.CreateProposal(c.Request.Context(), req, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Proposal created successfully",
		"data":    proposal,
	})
}

func (h *ProposalHandler) RejectProposal(c *gin.Context) {
	proposalID := c.Param("id")

	// Extract user ID from context.
	// NOTE: Adjust "user_id" to whatever key your JWT/Auth middleware uses (e.g., "userID", "sub").
	reviewerID := c.GetString("user_id")

	if reviewerID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: user ID not found in context"})
		return
	}

	err := h.proposalService.RejectProposal(c.Request.Context(), proposalID, reviewerID)
	if err != nil {
		// Basic error mapping.
		// In production, you might want to check if err contains "not found" to return 404 instead of 400.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Proposal rejected successfully",
		"data": map[string]interface{}{
			"id": proposalID,
		},
	})
}
