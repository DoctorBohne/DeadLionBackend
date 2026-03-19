package task_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	repo "github.com/DoctorBohne/DeadLionBackend/internal/repositories/deadline_objects"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func subtaskCols() []string {
	return []string{"id", "task_id", "title", "description", "board_pool", "created_at", "updated_at"}
}

func newSubtask(taskID uuid.UUID) *models.Subtask {
	desc := "a subtask"
	return &models.Subtask{
		ID:          uuid.New(),
		TaskID:      taskID,
		Title:       "Test Subtask",
		Description: &desc,
		BoardPool:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestSubtaskCreate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	taskID := uuid.New()
	// ownership COUNT
	mock.ExpectQuery(`SELECT count\(\*\) FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// INSERT
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "subtasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	s := newSubtask(taskID)
	if err := r.Create(context.Background(), 1, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSubtaskCreate_TaskNotOwned(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	// COUNT returns 0
	mock.ExpectQuery(`SELECT count\(\*\) FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	s := newSubtask(uuid.New())
	if err := r.Create(context.Background(), 1, s); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestSubtaskCreate_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "subtasks"`).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	s := newSubtask(uuid.New())
	if err := r.Create(context.Background(), 1, s); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── ListByTaskID ──────────────────────────────────────────────────────────────

func TestSubtaskListByTaskID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	taskID := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "subtasks"`).
		WillReturnRows(sqlmock.NewRows(subtaskCols()).
			AddRow(uuid.New(), taskID, "Sub A", nil, 0, now, now).
			AddRow(uuid.New(), taskID, "Sub B", nil, 1, now, now))

	items, err := r.ListByTaskID(context.Background(), 1, taskID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSubtaskListByTaskID_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "subtasks"`).
		WillReturnRows(sqlmock.NewRows(subtaskCols()))

	items, err := r.ListByTaskID(context.Background(), 1, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0, got %d", len(items))
	}
}

func TestSubtaskListByTaskID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "subtasks"`).
		WillReturnError(errors.New("db down"))

	_, err := r.ListByTaskID(context.Background(), 1, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestSubtaskGetByID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	id := uuid.New()
	taskID := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "subtasks"`).
		WillReturnRows(sqlmock.NewRows(subtaskCols()).
			AddRow(id, taskID, "Sub A", nil, 0, now, now))

	s, err := r.GetByID(context.Background(), 1, id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != id {
		t.Fatalf("expected ID %v, got %v", id, s.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSubtaskGetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "subtasks"`).
		WillReturnRows(sqlmock.NewRows(subtaskCols()))

	_, err := r.GetByID(context.Background(), 1, uuid.New())
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestSubtaskDelete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "subtasks"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	deleted, err := r.Delete(context.Background(), 1, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted=true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSubtaskDelete_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "subtasks"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	deleted, err := r.Delete(context.Background(), 1, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Fatal("expected deleted=false (not found)")
	}
}

func TestSubtaskDelete_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewSubtaskRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "subtasks"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	_, err := r.Delete(context.Background(), 1, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
