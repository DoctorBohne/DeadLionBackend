package handler

import (
	"context"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/services"
)

type RiskService interface {
	CalculateRiskList(ctx context.Context, userID uint, requestDate time.Time) ([]models.RiskItem, error)
}

type RiskRequest struct {
	RequestDate time.Time `form:"requestDate" json:"requestDate,omitempty" time_format:"2006-01-02"`
}

type UserService interface {
	FindOrCreate(ctx context.Context, in services.CreateUserInput) (*models.User, bool, error)
}

type MeUserService interface {
	FindOrCreate(ctx context.Context, in services.CreateUserInput) (*models.User, bool, error)
	MarkOnboardingComplete(ctx context.Context, issuer, sub string) error
}
