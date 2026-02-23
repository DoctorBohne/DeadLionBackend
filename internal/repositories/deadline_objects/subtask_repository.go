package task

import (
	"context"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubtaskRepo struct {
	db *gorm.DB
}

func NewSubtaskRepo(db *gorm.DB) *SubtaskRepo {
	return &SubtaskRepo{db: db}
}

// scoped limits all subtask queries to subtasks of tasks owned by the user.
func (r *SubtaskRepo) scoped(ctx context.Context, userID uint) *gorm.DB {
	// IMPORTANT: assumes tables are "subtasks" and "tasks"
	return r.db.WithContext(ctx).
		Model(&models.Subtask{}).
		Joins("JOIN tasks ON tasks.id = subtasks.task_id").
		Where("tasks.user_id = ?", userID)
}

func (r *SubtaskRepo) Create(ctx context.Context, userID uint, s *models.Subtask) error {
	// ensure parent task exists and belongs to user
	var cnt int64
	if err := r.db.WithContext(ctx).
		Table("tasks").
		Where("id = ? AND user_id = ?", s.TaskID, userID).
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return gorm.ErrRecordNotFound
	}

	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SubtaskRepo) ListByTaskID(ctx context.Context, userID uint, taskID uuid.UUID) ([]models.Subtask, error) {
	var items []models.Subtask
	err := r.scoped(ctx, userID).
		Where("subtasks.task_id = ?", taskID).
		Order("subtasks.created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *SubtaskRepo) GetByID(ctx context.Context, userID uint, taskID, id uuid.UUID) (*models.Subtask, error) {
	var s models.Subtask
	err := r.scoped(ctx, userID).
		Where("subtasks.task_id = ? AND subtasks.id = ?", taskID, id).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubtaskRepo) Update(ctx context.Context, userID uint, taskID, id uuid.UUID, updates map[string]any) (*models.Subtask, error) {
	tx := r.scoped(ctx, userID).
		Where("subtasks.task_id = ? AND subtasks.id = ?", taskID, id)

	res := tx.Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetByID(ctx, userID, taskID, id)
}

func (r *SubtaskRepo) Delete(ctx context.Context, userID uint, taskID, id uuid.UUID) (bool, error) {
	tx := r.scoped(ctx, userID).
		Where("subtasks.task_id = ? AND subtasks.id = ?", taskID, id)

	res := tx.Delete(&models.Subtask{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
