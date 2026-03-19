package services

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"gorm.io/gorm"
)

// stubUserLookup is a shared test double for the UserLookup interface.
type stubUserLookup struct {
	user *models.User
	err  error
}

func (s *stubUserLookup) FindByIssuerSub(_ context.Context, _, _ string) (*models.User, error) {
	return s.user, s.err
}

// userWithID returns a minimal *models.User with the given primary-key ID.
func userWithID(id uint) *models.User {
	u := &models.User{}
	u.Model = gorm.Model{ID: id}
	return u
}
