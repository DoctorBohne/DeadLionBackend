package services

import (
	"context"
	"errors"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/http/handler"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/repositories/user"
)

type UserRepo interface {
	FindByIssuerSub(ctx context.Context, issuer, sub string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	MarkOnboardingCompleted(ctx context.Context, issuer, sub string) error
	UpdateNameFieldsByIssuerSub(ctx context.Context, issuer, sub string, name, preferredUsername, givenName, familyName *string) error
	UpdateEmailByIssuerSub(ctx context.Context, issuer, sub string, email string, verified *bool) error
	SetEmailVerifiedByIssuerSub(ctx context.Context, issuer, sub string, verified bool) error
	UpdatePreferredUsernameByIssuerSub(ctx context.Context, issuer, sub, preferredUsername string) error
	UpdateGivenAndFamilyNameByIssuerSub(ctx context.Context, issuer, sub, givenName, familyName string) error
}
type UserService struct {
	r UserRepo
}

func NewUserService(r *user.Repo) *UserService {
	return &UserService{r}
}

func (u UserService) FindOrCreate(ctx context.Context, in handler.CreateUserInput) (*models.User, bool, error) {
	user, err := u.r.FindByIssuerSub(ctx, in.Issuer, in.Subject)
	if err == nil {
		return user, false, nil
	}
	if !errors.Is(err, custom_errors.ErrNotFound) {
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
func (s *UserService) MarkOnboardingComplete(ctx context.Context, issuer, sub string) error {
	return s.r.MarkOnboardingCompleted(ctx, issuer, sub)
}
func (s *UserService) UpdateNames(ctx context.Context, issuer, sub string, name, preferredUsername, givenName, familyName *string) error {
	return s.r.(UserRepo).UpdateNameFieldsByIssuerSub(ctx, issuer, sub, name, preferredUsername, givenName, familyName)
}

func (s *UserService) UpdateEmail(ctx context.Context, issuer, sub, email string, verified *bool) error {
	return s.r.(UserRepo).UpdateEmailByIssuerSub(ctx, issuer, sub, email, verified)
}

func (s *UserService) SetEmailVerified(ctx context.Context, issuer, sub string, verified bool) error {
	return s.r.(UserRepo).SetEmailVerifiedByIssuerSub(ctx, issuer, sub, verified)
}

func (s *UserService) UpdatePreferredUsername(ctx context.Context, issuer, sub, preferredUsername string) error {
	return s.r.(UserRepo).UpdatePreferredUsernameByIssuerSub(ctx, issuer, sub, preferredUsername)
}

func (s *UserService) UpdateGivenAndFamilyName(ctx context.Context, issuer, sub, givenName, familyName string) error {
	return s.r.(UserRepo).UpdateGivenAndFamilyNameByIssuerSub(ctx, issuer, sub, givenName, familyName)
}
