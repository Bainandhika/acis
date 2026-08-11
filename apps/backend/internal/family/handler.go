package family

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

func (h *Handler) CreateFamily(c *gin.Context) {
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

func (h *Handler) JoinFamily(c *gin.Context) {
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

func (h *Handler) GetMyFamily(c *gin.Context) {
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

func (h *Handler) CreateWallet(c *gin.Context) {
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

func (h *Handler) GetWallets(c *gin.Context) {
	familyID, _ := c.Get("family_id")
	fIDStr, _ := familyID.(string)

	res, err := h.svc.GetWallets(c.Request.Context(), fIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": res})
}
