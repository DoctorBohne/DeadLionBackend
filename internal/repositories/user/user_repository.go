package user

import (
	"context"
	"errors"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *Repo {
	return &Repo{db}
}

func (r *Repo) FindByIssuerSub(ctx context.Context, issuer, sub string) (*models.User, error) {
	var us models.User
	err := r.DB.WithContext(ctx).
		Where("issuer = ? AND sub = ?", issuer, sub).
		First(&us).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return &us, nil
}

func (r *Repo) Create(ctx context.Context, user *models.User) error {
	return r.DB.WithContext(ctx).Create(user).Error
}

func (r *Repo) MarkOnboardingCompleted(ctx context.Context, issuer, sub string) error {
	res := r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("issuer = ? AND sub = ? AND onboarding_completed = ?", issuer, sub, false).
		Update("onboarding_completed", true)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 1 {
		return nil
	}

	var u models.User
	err := r.DB.WithContext(ctx).
		Select("onboarding_completed").
		Where("issuer = ? AND sub = ?", issuer, sub).
		First(&u).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return custom_errors.ErrNotFound
		}
		return err
	}

	if u.OnboardingCompleted {
		return errors.New("onboarding already completed")
	}

	return errors.New("unexpected state: onboarding not updated")
}

func (r *Repo) Update(ctx context.Context, u *models.User) error {
	return r.DB.WithContext(ctx).Save(u).Error
}

func (r *Repo) UpdateNameFieldsByIssuerSub(
	ctx context.Context,
	issuer, sub string,
	name, preferredUsername, givenName, familyName *string,
) error {
	updates := map[string]any{}

	if name != nil {
		if v := strings.TrimSpace(*name); v != "" {
			updates["name"] = v
		}
	}
	if preferredUsername != nil {
		if v := strings.TrimSpace(*preferredUsername); v != "" {
			updates["preferred_username"] = v
		}
	}
	if givenName != nil {
		if v := strings.TrimSpace(*givenName); v != "" {
			updates["given_name"] = v
		}
	}
	if familyName != nil {
		if v := strings.TrimSpace(*familyName); v != "" {
			updates["family_name"] = v
		}
	}

	if len(updates) == 0 {
		return nil
	}

	res := r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("issuer = ? AND sub = ?", issuer, sub).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return custom_errors.ErrNotFound
	}
	return nil
}

func (r *Repo) UpdateEmailByIssuerSub(
	ctx context.Context,
	issuer, sub string,
	email string,
	verified *bool,
) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return errors.New("email must not be empty")
	}

	updates := map[string]any{"email": email}
	if verified != nil {
		updates["email_verified"] = *verified
	}

	res := r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("issuer = ? AND sub = ?", issuer, sub).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return custom_errors.ErrNotFound
	}
	return nil
}

func (r *Repo) SetEmailVerifiedByIssuerSub(ctx context.Context, issuer, sub string, verified bool) error {
	res := r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("issuer = ? AND sub = ?", issuer, sub).
		Update("email_verified", verified)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return custom_errors.ErrNotFound
	}
	return nil
}

func (r *Repo) UpdatePreferredUsernameByIssuerSub(ctx context.Context, issuer, sub, preferredUsername string) error {
	preferredUsername = strings.TrimSpace(preferredUsername)
	if preferredUsername == "" {
		return errors.New("preferred username must not be empty")
	}

	res := r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("issuer = ? AND sub = ?", issuer, sub).
		Update("preferred_username", preferredUsername)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return custom_errors.ErrNotFound
	}
	return nil
}

func (r *Repo) UpdateGivenAndFamilyNameByIssuerSub(ctx context.Context, issuer, sub, givenName, familyName string) error {
	updates := map[string]any{}
	if v := strings.TrimSpace(givenName); v != "" {
		updates["given_name"] = v
	}
	if v := strings.TrimSpace(familyName); v != "" {
		updates["family_name"] = v
	}
	if len(updates) == 0 {
		return nil
	}

	res := r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("issuer = ? AND sub = ?", issuer, sub).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return custom_errors.ErrNotFound
	}
	return nil
}

func (r *Repo) FindOrCreateByIssuerSub(ctx context.Context, user *models.User) (*models.User, bool, error) {
	res := r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "issuer"},
				{Name: "sub"},
			},
			DoNothing: true,
		}).
		Create(user)

	if res.Error != nil {
		return nil, false, res.Error
	}
	if res.RowsAffected == 1 {
		return user, true, nil
	}

	var existing models.User
	err := r.DB.WithContext(ctx).
		Where("issuer = ? AND sub = ?", user.Issuer, user.Sub).
		First(&existing).Error
	if err != nil {
		return nil, false, err
	}

	return &existing, false, nil
}
