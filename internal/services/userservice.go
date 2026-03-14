package services

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/repositories/user"
)

// CreateUserInput belongs to the service layer (NOT http/handler), to avoid import cycles.
type CreateUserInput struct {
	Issuer            string
	Subject           string
	Email             string
	EmailVerified     bool
	Name              string
	PreferredUsername string
	GivenName         string
	FamilyName        string
}

type UserRepo interface {
	FindByIssuerSub(ctx context.Context, issuer, sub string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	MarkOnboardingCompleted(ctx context.Context, issuer, sub string) error
	UpdateNameFieldsByIssuerSub(ctx context.Context, issuer, sub string, name, preferredUsername, givenName, familyName *string) error
	UpdateEmailByIssuerSub(ctx context.Context, issuer, sub string, email string, verified *bool) error
	SetEmailVerifiedByIssuerSub(ctx context.Context, issuer, sub string, verified bool) error
	UpdatePreferredUsernameByIssuerSub(ctx context.Context, issuer, sub, preferredUsername string) error
	UpdateGivenAndFamilyNameByIssuerSub(ctx context.Context, issuer, sub, givenName, familyName string) error
	FindOrCreateByIssuerSub(ctx context.Context, user *models.User) (*models.User, bool, error)
}

type UserService struct {
	r UserRepo
}

func NewUserService(r user.Repo) *UserService {
	return &UserService{r: &r}
}

func (s *UserService) FindOrCreate(ctx context.Context, in CreateUserInput) (*models.User, bool, error) {
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

	userR, created, err := s.r.FindOrCreateByIssuerSub(ctx, userC)
	if err != nil {
		return nil, false, err
	}
	return userR, created, err
}

func (s *UserService) MarkOnboardingComplete(ctx context.Context, issuer, sub string) error {
	return s.r.MarkOnboardingCompleted(ctx, issuer, sub)
}

func (s *UserService) UpdateNames(ctx context.Context, issuer, sub string, name, preferredUsername, givenName, familyName *string) error {
	return s.r.UpdateNameFieldsByIssuerSub(ctx, issuer, sub, name, preferredUsername, givenName, familyName)
}

func (s *UserService) UpdateEmail(ctx context.Context, issuer, sub, email string, verified *bool) error {
	return s.r.UpdateEmailByIssuerSub(ctx, issuer, sub, email, verified)
}

func (s *UserService) SetEmailVerified(ctx context.Context, issuer, sub string, verified bool) error {
	return s.r.SetEmailVerifiedByIssuerSub(ctx, issuer, sub, verified)
}

func (s *UserService) UpdatePreferredUsername(ctx context.Context, issuer, sub, preferredUsername string) error {
	return s.r.UpdatePreferredUsernameByIssuerSub(ctx, issuer, sub, preferredUsername)
}

func (s *UserService) UpdateGivenAndFamilyName(ctx context.Context, issuer, sub, givenName, familyName string) error {
	return s.r.UpdateGivenAndFamilyNameByIssuerSub(ctx, issuer, sub, givenName, familyName)
}
