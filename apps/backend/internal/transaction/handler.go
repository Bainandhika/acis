package transaction

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)

	var req CreateTransactionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserID = uidStr
	res, err := h.svc.CreateDirectTransaction(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Transaction created successfully",
		"data":    res,
	})
}

func (h *Handler) GetTransactions(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	res, err := h.svc.GetTransactions(c.Request.Context(), fIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *Handler) CreateProposal(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)

	var req CreateProposalDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ProposedBy = uidStr
	res, err := h.svc.CreateProposal(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Proposal created successfully",
		"data":    res,
	})
}

func (h *Handler) ApproveProposal(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)
	proposalID := c.Param("id")

	if err := h.svc.ApproveProposal(c.Request.Context(), proposalID, uidStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposal approved successfully"})
}

func (h *Handler) RejectProposal(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)
	proposalID := c.Param("id")

	if err := h.svc.RejectProposal(c.Request.Context(), proposalID, uidStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposal rejected successfully"})
}
