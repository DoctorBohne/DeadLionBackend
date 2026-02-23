package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepo interface {
	Create(ctx context.Context, t *models.Task) error
	ListByUserID(ctx context.Context, userID uint) ([]models.Task, error)
	ListByUserIDAndBoardPool(ctx context.Context, userID uint, boardPool int) ([]models.Task, error)
	GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Task, error)
	Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Task, error)
	Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

type CreateTaskInput struct {
	Issuer          string
	Subject         string
	Title           string
	Description     *string
	Risk            *string
	Priority        *string
	PriorityRank    *string
	IsFinished      *string // stored as string in model
	BoardPool       int
	EstimateMinutes int
	SpendMinutes    int
	DueAt           time.Time
}

type UpdateTaskInput struct {
	Title           *string
	Description     *string
	Risk            *string
	Priority        *string
	PriorityRank    *string
	IsFinished      *string
	BoardPool       *int
	EstimateMinutes *int
	SpendMinutes    *int
	DueAt           *time.Time
}

type TaskService interface {
	Create(ctx context.Context, in CreateTaskInput) (*models.Task, error)
	List(ctx context.Context, issuer, sub string, boardPool *int) ([]models.Task, error)
	GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Task, error)
	Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateTaskInput) (*models.Task, error)
	Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error
}

type taskService struct {
	users UserLookup
	repo  TaskRepo
}

func NewTaskService(repo TaskRepo, users UserLookup) TaskService {
	return &taskService{users: users, repo: repo}
}

func (s *taskService) userID(ctx context.Context, issuer, sub string) (uint, error) {
	u, err := s.users.FindByIssuerSub(ctx, issuer, sub)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, custom_errors.ErrNotFound
		}
		return 0, err
	}
	return u.ID, nil
}

func normalizeNonEmpty(field, v string) (string, error) {
	t := strings.TrimSpace(v)
	if t == "" {
		return "", errors.New(field + " must not be empty")
	}
	return t, nil
}

func normalizeBoolString(v string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(v))
	switch t {
	case "true", "false":
		return t, nil
	case "1":
		return "true", nil
	case "0":
		return "false", nil
	default:
		return "", errors.New("isFinished must be 'true' or 'false'")
	}
}

func validateMinutes(field string, n int) error {
	if n < 0 {
		return errors.New(field + " must be >= 0")
	}
	if n > 1_000_000 {
		return errors.New(field + " too large: " + strconv.Itoa(n))
	}
	return nil
}

func (s *taskService) Create(ctx context.Context, in CreateTaskInput) (*models.Task, error) {
	uid, err := s.userID(ctx, in.Issuer, in.Subject)
	if err != nil {
		return nil, err
	}

	title, err := normalizeNonEmpty("title", in.Title)
	if err != nil {
		return nil, err
	}

	if in.BoardPool < 0 {
		return nil, errors.New("boardPool must be >= 0")
	}
	if err := validateMinutes("estimateMinutes", in.EstimateMinutes); err != nil {
		return nil, err
	}
	if err := validateMinutes("spendMinutes", in.SpendMinutes); err != nil {
		return nil, err
	}
	if in.DueAt.IsZero() {
		return nil, errors.New("dueAt must be set")
	}

	t := &models.Task{
		UserID:          uid,
		Title:           title,
		Description:     in.Description,
		BoardPool:       in.BoardPool,
		EstimateMinutes: in.EstimateMinutes,
		SpendMinutes:    in.SpendMinutes,
		DueAt:           in.DueAt,
	}

	if in.Risk != nil {
		r, err := normalizeNonEmpty("risk", *in.Risk)
		if err != nil {
			return nil, err
		}
		t.Risk = strings.ToLower(r)
	}

	if in.Priority != nil {
		p, err := normalizeNonEmpty("priority", *in.Priority)
		if err != nil {
			return nil, err
		}
		t.Priority = p
	}

	if in.PriorityRank != nil {
		pr, err := normalizeNonEmpty("priorityRank", *in.PriorityRank)
		if err != nil {
			return nil, err
		}
		t.PriorityRank = pr
	}

	if in.IsFinished != nil {
		bs, err := normalizeBoolString(*in.IsFinished)
		if err != nil {
			return nil, err
		}
		t.IsFinished = bs
	} else {
		t.IsFinished = "false"
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *taskService) List(ctx context.Context, issuer, sub string, boardPool *int) ([]models.Task, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}
	if boardPool != nil {
		return s.repo.ListByUserIDAndBoardPool(ctx, uid, *boardPool)
	}
	return s.repo.ListByUserID(ctx, uid)
}

func (s *taskService) GetByID(ctx context.Context, issuer, sub string, id uuid.UUID) (*models.Task, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}

	t, err := s.repo.GetByID(ctx, uid, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *taskService) Update(ctx context.Context, issuer, sub string, id uuid.UUID, in UpdateTaskInput) (*models.Task, error) {
	uid, err := s.userID(ctx, issuer, sub)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}

	if in.Title != nil {
		t, err := normalizeNonEmpty("title", *in.Title)
		if err != nil {
			return nil, err
		}
		updates["title"] = t
	}

	if in.Description != nil {
		updates["description"] = in.Description
	}

	if in.Risk != nil {
		r, err := normalizeNonEmpty("risk", *in.Risk)
		if err != nil {
			return nil, err
		}
		updates["risk"] = strings.ToLower(r)
	}

	if in.Priority != nil {
		p, err := normalizeNonEmpty("priority", *in.Priority)
		if err != nil {
			return nil, err
		}
		updates["priority"] = p
	}

	if in.PriorityRank != nil {
		pr, err := normalizeNonEmpty("priorityRank", *in.PriorityRank)
		if err != nil {
			return nil, err
		}
		updates["priority_rank"] = pr
	}

	if in.IsFinished != nil {
		bs, err := normalizeBoolString(*in.IsFinished)
		if err != nil {
			return nil, err
		}
		updates["is_finished"] = bs
	}

	if in.BoardPool != nil {
		if *in.BoardPool < 0 {
			return nil, errors.New("boardPool must be >= 0")
		}
		updates["board_pool"] = *in.BoardPool
	}

	if in.EstimateMinutes != nil {
		if err := validateMinutes("estimateMinutes", *in.EstimateMinutes); err != nil {
			return nil, err
		}
		updates["estimate_minutes"] = *in.EstimateMinutes
	}

	if in.SpendMinutes != nil {
		if err := validateMinutes("spendMinutes", *in.SpendMinutes); err != nil {
			return nil, err
		}
		updates["spend_minutes"] = *in.SpendMinutes
	}

	if in.DueAt != nil {
		if in.DueAt.IsZero() {
			return nil, errors.New("dueAt must be set")
		}
		updates["due_at"] = *in.DueAt
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	t, err := s.repo.Update(ctx, uid, id, updates)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, custom_errors.ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *taskService) Delete(ctx context.Context, issuer, sub string, id uuid.UUID) error {
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
