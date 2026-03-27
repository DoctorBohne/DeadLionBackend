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

// ---- stub TaskboardRepo ----

type stubTaskboardRepo struct {
	createFn      func(ctx context.Context, userID uint, b *models.Taskboard) error
	listByTaskIDFn func(ctx context.Context, userID uint, taskID uuid.UUID) ([]models.Taskboard, error)
	getByIDFn     func(ctx context.Context, userID uint, id uuid.UUID) (*models.Taskboard, error)
	updateFn      func(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Taskboard, error)
	deleteFn      func(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

func (s *stubTaskboardRepo) Create(ctx context.Context, userID uint, b *models.Taskboard) error {
	return s.createFn(ctx, userID, b)
}
func (s *stubTaskboardRepo) ListByTaskID(ctx context.Context, userID uint, taskID uuid.UUID) ([]models.Taskboard, error) {
	return s.listByTaskIDFn(ctx, userID, taskID)
}
func (s *stubTaskboardRepo) GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Taskboard, error) {
	return s.getByIDFn(ctx, userID, id)
}
func (s *stubTaskboardRepo) Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Taskboard, error) {
	return s.updateFn(ctx, userID, id, updates)
}
func (s *stubTaskboardRepo) Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error) {
	return s.deleteFn(ctx, userID, id)
}

// ---- TaskboardService tests ----

func TestTaskboardService_Create_Success(t *testing.T) {
	taskID := uuid.New()
	repo := &stubTaskboardRepo{
		createFn: func(_ context.Context, _ uint, b *models.Taskboard) error { return nil },
	}
	svc := NewTaskboardService(repo, &stubUserLookup{user: userWithID(1)})
	board, err := svc.Create(context.Background(), CreateTaskboardInput{
		Issuer:  "iss",
		Subject: "sub",
		TaskID:  taskID,
		Title:   "Sprint Board",
	})
	if err != nil || board.Title != "Sprint Board" {
		t.Fatalf("unexpected: %v %v", board, err)
	}
}

func TestTaskboardService_Create_UserNotFound(t *testing.T) {
	svc := NewTaskboardService(&stubTaskboardRepo{}, &stubUserLookup{err: gorm.ErrRecordNotFound})
	_, err := svc.Create(context.Background(), CreateTaskboardInput{
		Issuer: "iss", Subject: "sub", Title: "X",
	})
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskboardService_Create_EmptyTitle(t *testing.T) {
	svc := NewTaskboardService(&stubTaskboardRepo{}, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Create(context.Background(), CreateTaskboardInput{
		Issuer: "iss", Subject: "sub", TaskID: uuid.New(), Title: "  ",
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestTaskboardService_List_Success(t *testing.T) {
	taskID := uuid.New()
	repo := &stubTaskboardRepo{
		listByTaskIDFn: func(_ context.Context, _ uint, _ uuid.UUID) ([]models.Taskboard, error) {
			return []models.Taskboard{{Title: "A"}, {Title: "B"}}, nil
		},
	}
	svc := NewTaskboardService(repo, &stubUserLookup{user: userWithID(1)})
	items, err := svc.List(context.Background(), "iss", "sub", taskID)
	if err != nil || len(items) != 2 {
		t.Fatalf("unexpected: %v %v", items, err)
	}
}

func TestTaskboardService_GetByID_Success(t *testing.T) {
	id := uuid.New()
	repo := &stubTaskboardRepo{
		getByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*models.Taskboard, error) {
			return &models.Taskboard{ID: id, Title: "T"}, nil
		},
	}
	svc := NewTaskboardService(repo, &stubUserLookup{user: userWithID(1)})
	board, err := svc.GetByID(context.Background(), "iss", "sub", id)
	if err != nil || board.ID != id {
		t.Fatalf("unexpected: %v %v", board, err)
	}
}

func TestTaskboardService_GetByID_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &stubTaskboardRepo{
		getByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*models.Taskboard, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewTaskboardService(repo, &stubUserLookup{user: userWithID(1)})
	_, err := svc.GetByID(context.Background(), "iss", "sub", id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskboardService_Delete_Success(t *testing.T) {
	id := uuid.New()
	repo := &stubTaskboardRepo{
		deleteFn: func(_ context.Context, _ uint, _ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := NewTaskboardService(repo, &stubUserLookup{user: userWithID(1)})
	if err := svc.Delete(context.Background(), "iss", "sub", id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskboardService_Delete_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &stubTaskboardRepo{
		deleteFn: func(_ context.Context, _ uint, _ uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	svc := NewTaskboardService(repo, &stubUserLookup{user: userWithID(1)})
	err := svc.Delete(context.Background(), "iss", "sub", id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
