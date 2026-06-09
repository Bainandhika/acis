package handler

import (
	"net/http"

	"github.com/Bainandhika/acis/apps/backend/internal/dto"
	"github.com/Bainandhika/acis/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type FamilyHandler struct {
	familyService service.FamilyService
}

// NewFamilyHandler creates a new instance of FamilyHandler
func NewFamilyHandler(familyService service.FamilyService) *FamilyHandler {
	return &FamilyHandler{familyService: familyService}
}

// CreateFamily handles POST /api/v1/families
func (h *FamilyHandler) CreateFamily(c *gin.Context) {
	var req dto.CreateFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	response, err := h.familyService.CreateFamily(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// JoinFamily handles POST /api/v1/families/join
func (h *FamilyHandler) JoinFamily(c *gin.Context) {
	var req dto.JoinFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	response, err := h.familyService.JoinFamily(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMyFamily handles GET /api/v1/families/me
func (h *FamilyHandler) GetMyFamily(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	response, err := h.familyService.GetMyFamily(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
