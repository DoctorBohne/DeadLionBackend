package boards

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserboardRepo interface {
	Create(ctx context.Context, b *models.Userboard) error
	ListByUserID(ctx context.Context, userID uint) ([]models.Userboard, error)
	GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Userboard, error)
	Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Userboard, error)
	Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

type GormRepo struct {
	db *gorm.DB
}

func NewUserboardRepo(db *gorm.DB) UserboardRepo {
	return &GormRepo{db: db}
}

func (r *GormRepo) Create(ctx context.Context, b *models.Userboard) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *GormRepo) ListByUserID(ctx context.Context, userID uint) ([]models.Userboard, error) {
	var out []models.Userboard
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&out).Error
	return out, err
}

func (r *GormRepo) GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Userboard, error) {
	var b models.Userboard
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *GormRepo) Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Userboard, error) {
	tx := r.db.WithContext(ctx).
		Model(&models.Userboard{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)

	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *GormRepo) Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error) {
	tx := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Userboard{})

	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}
