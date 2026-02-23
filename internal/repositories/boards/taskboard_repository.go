package boards

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskboardRepo struct {
	db *gorm.DB
}

func NewTaskboardRepo(db *gorm.DB) *TaskboardRepo {
	return &TaskboardRepo{db: db}
}

func (r *TaskboardRepo) scoped(ctx context.Context, userID uint) *gorm.DB {
	// IMPORTANT: assumes tasks table is "tasks" and has columns id (uuid) and user_id (uint)
	return r.db.WithContext(ctx).
		Model(&models.Taskboard{}).
		Joins("JOIN tasks ON tasks.id = taskboards.task_id").
		Where("tasks.user_id = ?", userID)
}

func (r *TaskboardRepo) Create(ctx context.Context, userID uint, b *models.Taskboard) error {
	// ensure task belongs to user (otherwise insert would leak existence)
	var cnt int64
	if err := r.db.WithContext(ctx).
		Table("tasks").
		Where("id = ? AND user_id = ?", b.TaskID, userID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *TaskboardRepo) ListByTaskID(ctx context.Context, userID uint, taskID uuid.UUID) ([]models.Taskboard, error) {
	var items []models.Taskboard
	err := r.scoped(ctx, userID).
		Where("taskboards.task_id = ?", taskID).
		Order("taskboards.created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *TaskboardRepo) GetByID(ctx context.Context, userID uint, taskID, id uuid.UUID) (*models.Taskboard, error) {
	var b models.Taskboard
	err := r.scoped(ctx, userID).
		Where("taskboards.task_id = ? AND taskboards.id = ?", taskID, id).
		First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *TaskboardRepo) Update(ctx context.Context, userID uint, taskID, id uuid.UUID, updates map[string]any) (*models.Taskboard, error) {
	tx := r.scoped(ctx, userID).Where("taskboards.task_id = ? AND taskboards.id = ?", taskID, id)

	res := tx.Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, userID, taskID, id)
}

func (r *TaskboardRepo) Delete(ctx context.Context, userID uint, taskID, id uuid.UUID) (bool, error) {
	tx := r.scoped(ctx, userID).Where("taskboards.task_id = ? AND taskboards.id = ?", taskID, id)
	res := tx.Delete(&models.Taskboard{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
