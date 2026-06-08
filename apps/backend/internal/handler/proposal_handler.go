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
