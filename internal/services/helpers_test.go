package services

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
)

// stubUserLookup is a shared test double for the UserLookup interface.
type stubUserLookup struct {
	user *models.User
	err  error
}

func (s *stubUserLookup) FindByIssuerSub(_ context.Context, _, _ string) (*models.User, error) {
	return s.user, s.err
}
