package services

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
)

type UserLookup interface {
	FindByIssuerSub(ctx context.Context, issuer, sub string) (*models.User, error)
}
