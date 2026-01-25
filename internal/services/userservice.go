package services

import (
	"context"
	"errors"

	"github.com/DoctorBohne/DeadLionBackend/internal/http/handler"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/repositories/user"
)

type UserRepo interface {
	FindByIssuerSub(ctx context.Context, issuer, sub string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}
type UserService struct {
	r UserRepo
}

func (u UserService) FindOrCreate(ctx context.Context, in handler.CreateUserInput) (*models.User, bool, error) {
	user, err := u.r.FindByIssuerSub(ctx, in.Issuer, in.Subject)
	if err == nil {
		return user, false, nil
	}
	if !errors.Is(err, user.ErrNotFound) {
		return nil, false, err
	}

	userC := &models.User{
		Sub:               in.Subject,
		Issuer:            in.Issuer,
		Email:             in.Email,
		EmailVerified:     in.EmailVerified,
		Name:              in.Name,
		PreferredUsername: in.PreferredUsername,
		GivenName:         in.GivenName,
		FamilyName:        in.FamilyName,
	}

	err1 := u.r.Create(ctx, userC)
	if err1 != nil {
		return nil, false, err
	}
	return userC, true, nil
}
