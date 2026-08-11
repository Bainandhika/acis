package family

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FamilyHandler struct {
	svc FamilyService
}

func NewHandler(svc FamilyService) *FamilyHandler {
	return &FamilyHandler{svc: svc}
}

func (h *FamilyHandler) CreateFamily(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, ok := userID.(string)
	if !ok || uidStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateFamilyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.CreateFamily(c.Request.Context(), uidStr, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Family created successfully",
		"data":    res,
	})
}

func (h *FamilyHandler) JoinFamily(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, ok := userID.(string)
	if !ok || uidStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req JoinFamilyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.JoinFamily(c.Request.Context(), uidStr, req.InviteCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Joined family successfully",
		"data":    res,
	})
}

func (h *FamilyHandler) GetMyFamily(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, ok := userID.(string)
	if !ok || uidStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	res, err := h.svc.GetMyFamily(c.Request.Context(), uidStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *FamilyHandler) CreateWallet(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, _ := userID.(string)
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	var req CreateWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.CreateWallet(c.Request.Context(), uidStr, fIDStr, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Wallet created successfully",
		"data":    res,
	})
}

func (h *FamilyHandler) GetWallets(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	res, err := h.svc.GetWallets(c.Request.Context(), fIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}
