package task

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepo struct {
	db *gorm.DB
}

func NewTaskRepo(db *gorm.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, t *models.Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TaskRepo) ListByUserID(ctx context.Context, userID uint) ([]models.Task, error) {
	var items []models.Task
	err := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ?", userID).
		Order("due_at ASC").
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

// Optional filter: list tasks by user + boardpool
func (r *TaskRepo) ListByUserIDAndBoardPool(ctx context.Context, userID uint, boardPool int) ([]models.Task, error) {
	var items []models.Task
	err := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ? AND board_pool = ?", userID, boardPool).
		Order("due_at ASC").
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *TaskRepo) GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Task, error) {
	var t models.Task
	err := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ? AND id = ?", userID, id).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Task, error) {
	tx := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("user_id = ? AND id = ?", userID, id)

	res := tx.Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, userID, id)
}

func (r *TaskRepo) Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error) {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		Delete(&models.Task{})

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
