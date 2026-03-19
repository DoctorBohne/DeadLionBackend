package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---- stub TaskRepo ----

type stubTaskRepo struct {
	createFn                   func(ctx context.Context, t *models.Task) error
	listByUserIDFn             func(ctx context.Context, userID uint) ([]models.Task, error)
	listByUserIDAndBoardPoolFn func(ctx context.Context, userID uint, boardPool int) ([]models.Task, error)
	getByIDFn                  func(ctx context.Context, userID uint, id uuid.UUID) (*models.Task, error)
	updateFn                   func(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Task, error)
	deleteFn                   func(ctx context.Context, userID uint, id uuid.UUID) (bool, error)
}

func (s *stubTaskRepo) Create(ctx context.Context, t *models.Task) error {
	return s.createFn(ctx, t)
}
func (s *stubTaskRepo) ListByUserID(ctx context.Context, userID uint) ([]models.Task, error) {
	return s.listByUserIDFn(ctx, userID)
}
func (s *stubTaskRepo) ListByUserIDAndBoardPool(ctx context.Context, userID uint, boardPool int) ([]models.Task, error) {
	return s.listByUserIDAndBoardPoolFn(ctx, userID, boardPool)
}
func (s *stubTaskRepo) GetByID(ctx context.Context, userID uint, id uuid.UUID) (*models.Task, error) {
	return s.getByIDFn(ctx, userID, id)
}
func (s *stubTaskRepo) Update(ctx context.Context, userID uint, id uuid.UUID, updates map[string]any) (*models.Task, error) {
	return s.updateFn(ctx, userID, id, updates)
}
func (s *stubTaskRepo) Delete(ctx context.Context, userID uint, id uuid.UUID) (bool, error) {
	return s.deleteFn(ctx, userID, id)
}

func userWithID(id uint) *models.User {
	u := &models.User{}
	u.ID = id
	return u
}

// ---- TaskService.Create tests ----

func TestTaskService_Create_Success(t *testing.T) {
	repo := &stubTaskRepo{
		createFn: func(_ context.Context, t *models.Task) error { return nil },
	}
	svc := NewTaskService(repo, &stubUserLookup{user: userWithID(1)})

	task, err := svc.Create(context.Background(), CreateTaskInput{
		Issuer:          "iss",
		Subject:         "sub",
		Title:           "New Task",
		BoardPool:       1,
		EstimateMinutes: 30,
		SpendMinutes:    5,
		DueAt:           time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "New Task" {
		t.Fatalf("unexpected title: %s", task.Title)
	}
}

func TestTaskService_Create_UserNotFound(t *testing.T) {
	svc := NewTaskService(&stubTaskRepo{}, &stubUserLookup{err: gorm.ErrRecordNotFound})

	_, err := svc.Create(context.Background(), CreateTaskInput{
		Issuer:  "iss",
		Subject: "sub",
	})
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Create_EmptyTitle(t *testing.T) {
	svc := NewTaskService(&stubTaskRepo{}, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Create(context.Background(), CreateTaskInput{
		Issuer:          "iss",
		Subject:         "sub",
		Title:           "   ",
		BoardPool:       1,
		EstimateMinutes: 10,
		SpendMinutes:    0,
		DueAt:           time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestTaskService_Create_RepoError(t *testing.T) {
	repo := &stubTaskRepo{
		createFn: func(_ context.Context, _ *models.Task) error { return errors.New("db error") },
	}
	svc := NewTaskService(repo, &stubUserLookup{user: userWithID(1)})
	_, err := svc.Create(context.Background(), CreateTaskInput{
		Issuer:          "iss",
		Subject:         "sub",
		Title:           "Task",
		BoardPool:       1,
		EstimateMinutes: 10,
		SpendMinutes:    5,
		DueAt:           time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ---- TaskService.List tests ----

func TestTaskService_List_AllTasks(t *testing.T) {
	repo := &stubTaskRepo{
		listByUserIDFn: func(_ context.Context, _ uint) ([]models.Task, error) {
			return []models.Task{{Title: "A"}, {Title: "B"}}, nil
		},
	}
	svc := NewTaskService(repo, &stubUserLookup{user: userWithID(1)})
	items, err := svc.List(context.Background(), "iss", "sub", nil)
	if err != nil || len(items) != 2 {
		t.Fatalf("unexpected: %v %v", items, err)
	}
}

func TestTaskService_List_WithBoardPool(t *testing.T) {
	bp := 3
	repo := &stubTaskRepo{
		listByUserIDAndBoardPoolFn: func(_ context.Context, _ uint, boardPool int) ([]models.Task, error) {
			if boardPool != 3 {
				return nil, errors.New("wrong boardPool")
			}
			return []models.Task{{Title: "X"}}, nil
		},
	}
	svc := NewTaskService(repo, &stubUserLookup{user: userWithID(1)})
	items, err := svc.List(context.Background(), "iss", "sub", &bp)
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected: %v %v", items, err)
	}
}

// ---- TaskService.GetByID tests ----

func TestTaskService_GetByID_Success(t *testing.T) {
	id := uuid.New()
	repo := &stubTaskRepo{
		getByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*models.Task, error) {
			return &models.Task{ID: id, Title: "T"}, nil
		},
	}
	svc := NewTaskService(repo, &stubUserLookup{user: userWithID(1)})
	task, err := svc.GetByID(context.Background(), "iss", "sub", id)
	if err != nil || task.ID != id {
		t.Fatalf("unexpected: %v %v", task, err)
	}
}

func TestTaskService_GetByID_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &stubTaskRepo{
		getByIDFn: func(_ context.Context, _ uint, _ uuid.UUID) (*models.Task, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	svc := NewTaskService(repo, &stubUserLookup{user: userWithID(1)})
	_, err := svc.GetByID(context.Background(), "iss", "sub", id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// ---- TaskService.Delete tests ----

func TestTaskService_Delete_Success(t *testing.T) {
	id := uuid.New()
	repo := &stubTaskRepo{
		deleteFn: func(_ context.Context, _ uint, _ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := NewTaskService(repo, &stubUserLookup{user: userWithID(1)})
	if err := svc.Delete(context.Background(), "iss", "sub", id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTaskService_Delete_NotFound(t *testing.T) {
	id := uuid.New()
	repo := &stubTaskRepo{
		deleteFn: func(_ context.Context, _ uint, _ uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	svc := NewTaskService(repo, &stubUserLookup{user: userWithID(1)})
	err := svc.Delete(context.Background(), "iss", "sub", id)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
