package user

import (
	"context"
	"errors"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"gorm.io/gorm"
)

type Repo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *Repo {
	return &Repo{db}
}

func (u Repo) FindByIssuerSub(ctx context.Context, issuer, sub string) (*models.User, error) {
	var us models.User
	err := u.DB.WithContext(ctx).
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

func (u Repo) Create(ctx context.Context, user *models.User) error {
	return u.DB.WithContext(ctx).Create(user).Error
}
