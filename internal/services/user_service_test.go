package services

import (
	"context"
	"errors"
	"testing"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
)

// ---- stub UserRepo ----

type stubUserRepo struct {
	findByIssuerSubFn                    func(ctx context.Context, issuer, sub string) (*models.User, error)
	createFn                             func(ctx context.Context, user *models.User) error
	markOnboardingCompletedFn            func(ctx context.Context, issuer, sub string) error
	findOrCreateByIssuerSubFn            func(ctx context.Context, user *models.User) (*models.User, bool, error)
	updateNameFieldsByIssuerSubFn         func(ctx context.Context, issuer, sub string, name, preferredUsername, givenName, familyName *string) error
	updateEmailByIssuerSubFn             func(ctx context.Context, issuer, sub string, email string, verified *bool) error
	setEmailVerifiedByIssuerSubFn        func(ctx context.Context, issuer, sub string, verified bool) error
	updatePreferredUsernameByIssuerSubFn  func(ctx context.Context, issuer, sub, preferredUsername string) error
	updateGivenAndFamilyNameByIssuerSubFn func(ctx context.Context, issuer, sub, givenName, familyName string) error
}

func (s *stubUserRepo) FindByIssuerSub(ctx context.Context, issuer, sub string) (*models.User, error) {
	return s.findByIssuerSubFn(ctx, issuer, sub)
}
func (s *stubUserRepo) Create(ctx context.Context, user *models.User) error {
	return s.createFn(ctx, user)
}
func (s *stubUserRepo) MarkOnboardingCompleted(ctx context.Context, issuer, sub string) error {
	return s.markOnboardingCompletedFn(ctx, issuer, sub)
}
func (s *stubUserRepo) FindOrCreateByIssuerSub(ctx context.Context, user *models.User) (*models.User, bool, error) {
	return s.findOrCreateByIssuerSubFn(ctx, user)
}
func (s *stubUserRepo) UpdateNameFieldsByIssuerSub(ctx context.Context, issuer, sub string, name, preferredUsername, givenName, familyName *string) error {
	return s.updateNameFieldsByIssuerSubFn(ctx, issuer, sub, name, preferredUsername, givenName, familyName)
}
func (s *stubUserRepo) UpdateEmailByIssuerSub(ctx context.Context, issuer, sub string, email string, verified *bool) error {
	return s.updateEmailByIssuerSubFn(ctx, issuer, sub, email, verified)
}
func (s *stubUserRepo) SetEmailVerifiedByIssuerSub(ctx context.Context, issuer, sub string, verified bool) error {
	return s.setEmailVerifiedByIssuerSubFn(ctx, issuer, sub, verified)
}
func (s *stubUserRepo) UpdatePreferredUsernameByIssuerSub(ctx context.Context, issuer, sub, preferredUsername string) error {
	return s.updatePreferredUsernameByIssuerSubFn(ctx, issuer, sub, preferredUsername)
}
func (s *stubUserRepo) UpdateGivenAndFamilyNameByIssuerSub(ctx context.Context, issuer, sub, givenName, familyName string) error {
	return s.updateGivenAndFamilyNameByIssuerSubFn(ctx, issuer, sub, givenName, familyName)
}

// ---- UserService tests ----

func TestUserService_FindOrCreate_Created(t *testing.T) {
	repo := &stubUserRepo{
		findOrCreateByIssuerSubFn: func(_ context.Context, u *models.User) (*models.User, bool, error) {
			return u, true, nil
		},
	}
	svc := UserService{r: repo}
	user, created, err := svc.FindOrCreate(context.Background(), CreateUserInput{
		Issuer:  "https://issuer.test",
		Subject: "sub-1",
		Email:   "test@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected created=true")
	}
	if user.Issuer != "https://issuer.test" {
		t.Fatalf("unexpected issuer: %s", user.Issuer)
	}
}

func TestUserService_FindOrCreate_Existing(t *testing.T) {
	repo := &stubUserRepo{
		findOrCreateByIssuerSubFn: func(_ context.Context, u *models.User) (*models.User, bool, error) {
			return u, false, nil
		},
	}
	svc := UserService{r: repo}
	_, created, err := svc.FindOrCreate(context.Background(), CreateUserInput{Issuer: "x", Subject: "y"})
	if err != nil || created {
		t.Fatalf("expected created=false, got created=%v err=%v", created, err)
	}
}

func TestUserService_FindOrCreate_RepoError(t *testing.T) {
	repo := &stubUserRepo{
		findOrCreateByIssuerSubFn: func(_ context.Context, _ *models.User) (*models.User, bool, error) {
			return nil, false, errors.New("db error")
		},
	}
	svc := UserService{r: repo}
	_, _, err := svc.FindOrCreate(context.Background(), CreateUserInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUserService_MarkOnboardingComplete_Success(t *testing.T) {
	repo := &stubUserRepo{
		markOnboardingCompletedFn: func(_ context.Context, _, _ string) error { return nil },
	}
	svc := UserService{r: repo}
	if err := svc.MarkOnboardingComplete(context.Background(), "iss", "sub"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_MarkOnboardingComplete_Error(t *testing.T) {
	repo := &stubUserRepo{
		markOnboardingCompletedFn: func(_ context.Context, _, _ string) error {
			return errors.New("already done")
		},
	}
	svc := UserService{r: repo}
	if err := svc.MarkOnboardingComplete(context.Background(), "iss", "sub"); err == nil {
		t.Fatal("expected error")
	}
}

func TestUserService_UpdateNames_Success(t *testing.T) {
	repo := &stubUserRepo{
		updateNameFieldsByIssuerSubFn: func(_ context.Context, _, _ string, _, _, _, _ *string) error { return nil },
	}
	svc := UserService{r: repo}
	if err := svc.UpdateNames(context.Background(), "iss", "sub", nil, nil, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_UpdateEmail_Success(t *testing.T) {
	repo := &stubUserRepo{
		updateEmailByIssuerSubFn: func(_ context.Context, _, _, _ string, _ *bool) error { return nil },
	}
	svc := UserService{r: repo}
	if err := svc.UpdateEmail(context.Background(), "iss", "sub", "a@b.com", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_SetEmailVerified_Success(t *testing.T) {
	repo := &stubUserRepo{
		setEmailVerifiedByIssuerSubFn: func(_ context.Context, _, _ string, _ bool) error { return nil },
	}
	svc := UserService{r: repo}
	if err := svc.SetEmailVerified(context.Background(), "iss", "sub", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_UpdatePreferredUsername_Success(t *testing.T) {
	repo := &stubUserRepo{
		updatePreferredUsernameByIssuerSubFn: func(_ context.Context, _, _, _ string) error { return nil },
	}
	svc := UserService{r: repo}
	if err := svc.UpdatePreferredUsername(context.Background(), "iss", "sub", "user123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_UpdateGivenAndFamilyName_Success(t *testing.T) {
	repo := &stubUserRepo{
		updateGivenAndFamilyNameByIssuerSubFn: func(_ context.Context, _, _, _, _ string) error { return nil },
	}
	svc := UserService{r: repo}
	if err := svc.UpdateGivenAndFamilyName(context.Background(), "iss", "sub", "Max", "Müller"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
