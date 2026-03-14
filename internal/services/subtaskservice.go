package services

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubtaskRepo interface {
	Create(ctx context.Context, userID uint, s *models.Subtask) error
	ListByTaskID(ctx context.Context, userID uint, taskID uuid.UUID) ([]models.Subtask, error)
	GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Subtask, error)
	Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Subtask, error)
	Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

type CreateSubtaskInput struct {
	Issuer      string
	Subject     string
	TaskID      uuid.UUID
	Title       string
	Description *string
	BoardPool   int
}

type UpdateSubtaskInput struct {
	Title       *string
	Description *string
	BoardPool   *int
}

type SubtaskService interface {
	Create(ctx context.Context, in CreateSubtaskInput) (*models.Subtask, error)
	List(ctx context.Context, issuer, sub string, taskID uuid.UUID) ([]models.Subtask, error)
	GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Subtask, error)
	Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateSubtaskInput) (*models.Subtask, error)
	Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

type subtaskService struct {
	users UserLookup
	repo  SubtaskRepo
}

func NewSubtaskService(repo SubtaskRepo, users UserLookup) SubtaskService {
	return &subtaskService{users: users, repo: repo}
}

func (s *subtaskService) userID(ctx context.Context, issuer, sub string) (uint, error) {
	u, err := s.users.FindByIssuerSub(ctx, issuer, sub)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, custom_errors.ErrNotFound
		}
		return 0, err
	}
	return u.ID, nil
}

func validateBoardPool(p int) error {
	if p < 0 {
		return errors.New("boardPool must be >= 0")
	}
	if p > 1_000_000 {
		return errors.New("boardPool too large: " + strconv.Itoa(p))
	}
	return nil
}

func (s *subtaskService) Create(ctx context.Context, in CreateSubtaskInput) (*models.Subtask, error) {
	uid, err := s.userID(ctx, in.Issuer, in.Subject)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, errors.New("title must not be empty")
	}
	if err := validateBoardPool(in.BoardPool); err != nil {
		return nil, err
	}

	st := &models.Subtask{
		TaskID:      in.TaskID,
		Title:       title,
		Description: in.Description,
		BoardPool:   in.BoardPool,
	}

	if err := s.repo.Create(ctx, uid, st); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// parent task not found OR not owned by user
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return st, nil
}

func (s *subtaskService) List(ctx context.Context, issuer, sub string, taskID uuid.UUID) ([]models.Subtask, error) {
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

func (s *subtaskService) GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Subtask, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}

	it, err := s.repo.GetByID(ctx, uid, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return it, nil
}

func (s *subtaskService) Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateSubtaskInput) (*models.Subtask, error) {
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
	if in.BoardPool != nil {
		if err := validateBoardPool(*in.BoardPool); err != nil {
			return nil, err
		}
		updates["board_pool"] = *in.BoardPool
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	it, err := s.repo.Update(ctx, uid, id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return it, nil
}

func (s *subtaskService) Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error {
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
