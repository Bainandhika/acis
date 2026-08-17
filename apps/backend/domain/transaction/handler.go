package transaction

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	svc TransactionService
}

func NewHandler(svc TransactionService) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
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

func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)
	txID := c.Param("id")

	var req UpdateTransactionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.UserID = uidStr
	req.FamilyID = fIDStr
	res, err := h.svc.UpdateTransaction(c.Request.Context(), txID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction updated successfully",
		"data":    res,
	})
}

func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	res, err := h.svc.GetTransactions(c.Request.Context(), fIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *TransactionHandler) CreateProposal(c *gin.Context) {
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

func (h *TransactionHandler) GetProposals(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	res, err := h.svc.GetProposals(c.Request.Context(), fIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *TransactionHandler) ApproveProposal(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)
	proposalID := c.Param("id")

	if err := h.svc.ApproveProposal(c.Request.Context(), proposalID, uidStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposal approved successfully"})
}

func (h *TransactionHandler) RejectProposal(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)
	proposalID := c.Param("id")

	if err := h.svc.RejectProposal(c.Request.Context(), proposalID, uidStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Proposal rejected successfully"})
}

func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)
	txID := c.Param("id")

	if err := h.svc.DeleteTransaction(c.Request.Context(), txID, fIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}
