package boards_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"github.com/DoctorBohne/DeadLionBackend/internal/repositories/boards"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func taskboardCols() []string {
	return []string{"id", "task_id", "title", "description", "status", "created_at", "updated_at"}
}

func newTaskboard(taskID uuid.UUID) *models.Taskboard {
	desc := "a taskboard"
	return &models.Taskboard{
		ID:          uuid.New(),
		TaskID:      taskID,
		Title:       "Test Taskboard",
		Description: &desc,
		Status:      "todo",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestTaskboardCreate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	taskID := uuid.New()
	// ownership COUNT
	mock.ExpectQuery(`SELECT count\(\*\) FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// INSERT
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "taskboards"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	b := newTaskboard(taskID)
	if err := r.Create(context.Background(), 1, b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskboardCreate_TaskNotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	// COUNT returns 0
	mock.ExpectQuery(`SELECT count\(\*\) FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	b := newTaskboard(uuid.New())
	err := r.Create(context.Background(), 1, b)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── ListByTaskID ──────────────────────────────────────────────────────────────

func TestTaskboardListByTaskID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	taskID := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "taskboards"`).
		WillReturnRows(sqlmock.NewRows(taskboardCols()).
			AddRow(uuid.New(), taskID, "Board 1", nil, "todo", now, now).
			AddRow(uuid.New(), taskID, "Board 2", nil, "done", now, now))

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

func TestTaskboardListByTaskID_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "taskboards"`).
		WillReturnRows(sqlmock.NewRows(taskboardCols()))

	items, err := r.ListByTaskID(context.Background(), 1, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0, got %d", len(items))
	}
}

func TestTaskboardListByTaskID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "taskboards"`).
		WillReturnError(errors.New("db down"))

	_, err := r.ListByTaskID(context.Background(), 1, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestTaskboardGetByID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	id := uuid.New()
	taskID := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "taskboards"`).
		WillReturnRows(sqlmock.NewRows(taskboardCols()).
			AddRow(id, taskID, "Board A", nil, "todo", now, now))

	b, err := r.GetByID(context.Background(), 1, id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ID != id {
		t.Fatalf("expected ID %v, got %v", id, b.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskboardGetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "taskboards"`).
		WillReturnRows(sqlmock.NewRows(taskboardCols()))

	_, err := r.GetByID(context.Background(), 1, uuid.New())
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestTaskboardUpdate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	id := uuid.New()
	taskID := uuid.New()
	now := time.Now()
	// UPDATE
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "taskboards"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// follow-up GetByID
	mock.ExpectQuery(`SELECT .* FROM "taskboards"`).
		WillReturnRows(sqlmock.NewRows(taskboardCols()).
			AddRow(id, taskID, "Updated", nil, "in_progress", now, now))

	b, err := r.Update(context.Background(), 1, id, map[string]any{"title": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Title != "Updated" {
		t.Fatalf("expected Updated, got %s", b.Title)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskboardUpdate_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "taskboards"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	_, err := r.Update(context.Background(), 1, uuid.New(), map[string]any{"title": "X"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestTaskboardDelete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "taskboards"`).
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

func TestTaskboardDelete_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "taskboards"`).
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

func TestTaskboardDelete_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewTaskboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "taskboards"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	_, err := r.Delete(context.Background(), 1, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
