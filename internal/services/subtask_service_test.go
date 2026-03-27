package services

import (
	"context"
	"errors"
	"testing"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---- stub SubtaskRepo ----

type stubSubtaskRepo struct {
	createFn      func(ctx context.Context, userID uint, s *models.Subtask) error
	listByTaskIDFn func(ctx context.Context, userID uint, taskID uuid.UUID) ([]models.Subtask, error)
	getByIDFn     func(ctx context.Context, userID uint, id uuid.UUID) (*models.Subtask, error)
	updateFn      func(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Subtask, error)
	deleteFn      func(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

func (s *stubSubtaskRepo) Create(ctx context.Context, userID uint, sub *models.Subtask) error {
	return s.createFn(ctx, userID, sub)
}
func (s *stubSubtaskRepo) ListByTaskID(ctx context.Context, userID uint, taskID uuid.UUID) ([]models.Subtask, error) {
	return s.listByTaskIDFn(ctx, userID, taskID)
}
func (s *stubSubtaskRepo) GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Subtask, error) {
	return s.getByIDFn(ctx, userID, id)
}
func (s *stubSubtaskRepo) Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Subtask, error) {
	return s.updateFn(ctx, userID, id, updates)
}
func (s *stubSubtaskRepo) Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error) {
	return s.deleteFn(ctx, userID, id)
}

// ---- SubtaskService tests ----

func TestSubtaskService_Create_Success(t *testing.T) {
	taskID := uuid.New()
	repo := &stubSubtaskRepo{
		createFn: func(_ context.Context, _ uint, _ *models.Subtask) error { return nil },
	}
	svc := NewSubtaskService(repo, &stubUserLookup{user: userWithID(1)})
	item, err := svc.Create(context.Background(), CreateSubtaskInput{
		Issuer:    "iss",
		Subject:   "sub",
		TaskID:    taskID,
		Title:     "Sub 1",
		BoardPool: 2,
	})
	if err != nil || item.Title != "Sub 1" {
		t.Fatalf("unexpected: %v %v", item, err)
	}
}

func TestSubtaskService_Create_UserNotFound(t *testing.T) {
	svc := NewSubtaskService(&stubSubtaskRepo{}, &stubUserLookup{err: gorm.ErrRecordNotFound})
	_, err := svc.Create(context.Background(), CreateSubtaskInput{
		Issuer: "iss", Subject: "sub", Title: "X",
	})
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSubtaskService_Create_EmptyTitle(t *testing.T) {
	svc := NewSubtaskService(&stubSubtaskRepo{}, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Create(context.Background(), CreateSubtaskInput{
		Issuer: "iss", Subject: "sub", TaskID: uuid.New(), Title: "  ", BoardPool: 1,
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestSubtaskService_List_Success(t *testing.T) {
	taskID := uuid.New()
	repo := &stubSubtaskRepo{
		listByTaskIDFn: func(_ context.Context, _ uint, _ uuid.UUID) ([]models.Subtask, error) {
			return []models.Subtask{{Title: "A"}, {Title: "B"}}, nil
		},
	}
	svc := NewSubtaskService(repo, &stubUserLookup{user: userWithID(1)})
	items, err := svc.List(context.Background(), "iss", "sub", taskID)
	if err != nil || len(items) != 2 {
		t.Fatalf("unexpected: %v %v", items, err)
	}
}

func TestSubtaskService_GetByID_Success(t *testing.T) {
	id := uuid.New()
	repo := &stubSubtaskRepo{
		getByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*models.Subtask, error) {
			return &models.Subtask{ID: id, Title: "S"}, nil
		},
	}
	svc := NewSubtaskService(repo, &stubUserLookup{user: userWithID(1)})
	item, err := svc.GetByID(context.Background(), "iss", "sub", id)
	if err != nil || item.ID != id {
		t.Fatalf("unexpected: %v %v", item, err)
	}
}

func TestSubtaskService_GetByID_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &stubSubtaskRepo{
		getByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*models.Subtask, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewSubtaskService(repo, &stubUserLookup{user: userWithID(1)})
	_, err := svc.GetByID(context.Background(), "iss", "sub", id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSubtaskService_Delete_Success(t *testing.T) {
	id := uuid.New()
	repo := &stubSubtaskRepo{
		deleteFn: func(_ context.Context, _ uint, _ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := NewSubtaskService(repo, &stubUserLookup{user: userWithID(1)})
	if err := svc.Delete(context.Background(), "iss", "sub", id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubtaskService_Delete_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &stubSubtaskRepo{
		deleteFn: func(_ context.Context, _ uint, _ uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	svc := NewSubtaskService(repo, &stubUserLookup{user: userWithID(1)})
	err := svc.Delete(context.Background(), "iss", "sub", id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
