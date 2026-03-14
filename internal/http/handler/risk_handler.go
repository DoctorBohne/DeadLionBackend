package handler

import (
	"net/http"

	"github.com/DoctorBohne/DeadLionBackend/internal/helper"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
	"github.com/gin-gonic/gin"
)

// RiskHandler Risk handler struct definition
type RiskHandler struct {
	risksvc RiskService
	usersvc services.UserLookup
}

// NewRiskHandler constructor
func NewRiskHandler(risksvc RiskService, userlk services.UserLookup) *RiskHandler {
	return &RiskHandler{risksvc: risksvc,
		usersvc: userlk}
}

func (h *RiskHandler) RetrieveRiskList(c *gin.Context) {
	issuer, sub, ok := helper.MustClaims(c)
	if !ok || issuer == "" || sub == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.usersvc.FindByIssuerSub(c.Request.Context(), issuer, sub)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := user.ID
	var req RiskRequest
	err = c.ShouldBindQuery(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestDate := req.RequestDate

	risklist, err := h.risksvc.CalculateRiskList(c.Request.Context(), userID, requestDate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"risklist": risklist})
}
