package handler

import (
	"context"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
)

type RiskService interface {
	CalculateRiskList(ctx context.Context, userID uint, requestDate time.Time) ([]models.RiskItem, error)
}

type RiskRequest struct {
	RequestDate time.Time `json:"requestDate,omitempty" time_format:"2006-01-02"`
}
