package handler

import (
	"net/http"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/requestctx"
	"github.com/gin-gonic/gin"
)

type AbgabeHandler struct {
}

func NewAbgabeHandler() *AbgabeHandler {
	return &AbgabeHandler{}
}

type CreateAbgabeRequest struct {
	Title          string      `json:"title"`
	DueDate        time.Time   `json:"due_date"`
	RiskAssessment abgabe.Risk `json:"risk_assessment"`
	ModulID        uint        `json:"modul_id"`
}

func (a *AbgabeHandler) CreateAbgabe(c *gin.Context, r *CreateAbgabeRequest) {
	claims, ok := requestctx.ClaimsFrom(c.Request.Context())
	if !ok || claims.Subject == "" || claims.Issuer == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

}
