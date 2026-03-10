package services

import (
	"context"
	"errors"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskboardRepo interface {
	Create(ctx context.Context, userID uint, b *models.Taskboard) error
	ListByTaskID(ctx context.Context, userID uint, taskID uuid.UUID) ([]models.Taskboard, error)
	GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Taskboard, error)
	Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Taskboard, error)
	Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

type CreateTaskboardInput struct {
	Issuer      string
	Subject     string
	TaskID      uuid.UUID
	Title       string
	Description *string
	Status      *string // optional, default in DB is 'todo'
}

type UpdateTaskboardInput struct {
	Title       *string
	Description *string
	Status      *string
}

type TaskboardService interface {
	Create(ctx context.Context, in CreateTaskboardInput) (*models.Taskboard, error)
	List(ctx context.Context, issuer, sub string, taskID uuid.UUID) ([]models.Taskboard, error)
	GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Taskboard, error)
	Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateTaskboardInput) (*models.Taskboard, error)
	Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

type taskboardService struct {
	users UserLookup
	repo  TaskboardRepo
}

func NewTaskboardService(repo TaskboardRepo, users UserLookup) TaskboardService {
	return &taskboardService{users: users, repo: repo}
}

func (s *taskboardService) userID(ctx context.Context, issuer, sub string) (uint, error) {
	u, err := s.users.FindByIssuerSub(ctx, issuer, sub)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, custom_errors.ErrNotFound
		}
		return 0, err
	}
	return u.ID, nil
}

func normalizeStatus(st string) string {
	return strings.ToLower(strings.TrimSpace(st))
}

func isAllowedStatus(st string) bool {
	switch st {
	case "todo", "doing", "done", "blocked":
		return true
	default:
		return false
	}
}

func (s *taskboardService) Create(ctx context.Context, in CreateTaskboardInput) (*models.Taskboard, error) {
	uid, err := s.userID(ctx, in.Issuer, in.Subject)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, errors.New("title must not be empty")
	}

	b := &models.Taskboard{
		TaskID:      in.TaskID,
		Title:       title,
		Description: in.Description,
	}

	if in.Status != nil {
		st := normalizeStatus(*in.Status)
		if st == "" {
			return nil, errors.New("status must not be empty")
		}
		if !isAllowedStatus(st) {
			return nil, errors.New("invalid status")
		}
		b.Status = st
	}

	if err := s.repo.Create(ctx, uid, b); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// either user not found (handled earlier) or task not found/not owned
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}

	return b, nil
}

func (s *taskboardService) List(ctx context.Context, issuer, sub string, taskID uuid.UUID) ([]models.Taskboard, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListByTaskID(ctx, uid, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return items, nil
}

func (s *taskboardService) GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Taskboard, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}

	b, err := s.repo.GetByID(ctx, uid, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *taskboardService) Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateTaskboardInput) (*models.Taskboard, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}

	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return nil, errors.New("title must not be empty")
		}
		updates["title"] = t
	}

	if in.Description != nil {
		updates["description"] = in.Description
	}

	if in.Status != nil {
		st := normalizeStatus(*in.Status)
		if st == "" {
			return nil, errors.New("status must not be empty")
		}
		if !isAllowedStatus(st) {
			return nil, errors.New("invalid status")
		}
		updates["status"] = st
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	b, err := s.repo.Update(ctx, uid, id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *taskboardService) Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return err
	}

	ok, err := s.repo.Delete(ctx, uid, id)
	if err != nil {
		return err
	}
	if !ok {
		return custom_errors.ErrNotFound
	}
	return nil
}
