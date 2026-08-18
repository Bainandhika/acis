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

	var income float64
	if req.MonthlyIncome != nil {
		income = *req.MonthlyIncome
	}
	res, err := h.svc.CreateFamily(c.Request.Context(), uidStr, req.Name, income)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res == nil {
		c.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}

func (h *FamilyHandler) UpdateSettings(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	var req UpdateFamilySettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateFamilySettings(c.Request.Context(), fIDStr, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Family settings updated successfully"})
}

func (h *FamilyHandler) DisconnectTelegram(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	if err := h.svc.DisconnectTelegram(c.Request.Context(), fIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Telegram disconnected successfully"})
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

func (h *FamilyHandler) UpdateFamily(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	var req UpdateFamilyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateFamilyName(c.Request.Context(), fIDStr, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Family name updated successfully"})
}

func (h *FamilyHandler) UpdateWallet(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)
	walletID := c.Param("id")

	var req UpdateWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.UpdateWallet(c.Request.Context(), walletID, fIDStr, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Wallet updated successfully",
		"data":    res,
	})
}

func (h *FamilyHandler) DeleteWallet(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)
	walletID := c.Param("id")

	if err := h.svc.DeleteWallet(c.Request.Context(), walletID, fIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wallet deleted successfully"})
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

func (h *FamilyHandler) RemoveMember(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uidStr, ok := userID.(string)
	if !ok || uidStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)
	memberID := c.Param("id")

	if err := h.svc.RemoveMember(c.Request.Context(), uidStr, memberID, fIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}
