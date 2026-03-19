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

func taskCols() []string {
	return []string{
		"id", "user_id", "title", "description",
		"risk", "priority", "priority_rank", "is_finished",
		"board_pool", "estimate_minutes", "spend_minutes",
		"due_at", "created_at", "updated_at",
	}
}

func newTask(userID uint) *models.Task {
	desc := "a task"
	return &models.Task{
		ID:              uuid.New(),
		UserID:          userID,
		Title:           "Test Task",
		Description:     &desc,
		Risk:            "low",
		Priority:        "normal",
		PriorityRank:    "1",
		IsFinished:      "false",
		BoardPool:       0,
		EstimateMinutes: 30,
		SpendMinutes:    0,
		DueAt:           time.Now().Add(24 * time.Hour),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestTaskCreate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tasks"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	task := newTask(1)
	if err := r.Create(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskCreate_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tasks"`).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	if err := r.Create(context.Background(), newTask(1)); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── ListByUserID ──────────────────────────────────────────────────────────────

func TestTaskListByUserID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	now := time.Now()
	due := now.Add(24 * time.Hour)
	mock.ExpectQuery(`SELECT .* FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows(taskCols()).
			AddRow(uuid.New(), 1, "Task A", nil, "low", "normal", "1", "false", 0, 30, 0, due, now, now).
			AddRow(uuid.New(), 1, "Task B", nil, "high", "high", "2", "false", 1, 60, 10, due, now, now))

	items, err := r.ListByUserID(context.Background(), 1)
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

func TestTaskListByUserID_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows(taskCols()))

	items, err := r.ListByUserID(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0, got %d", len(items))
	}
}

func TestTaskListByUserID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "tasks"`).
		WillReturnError(errors.New("db down"))

	_, err := r.ListByUserID(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── ListByUserIDAndBoardPool ──────────────────────────────────────────────────

func TestTaskListByUserIDAndBoardPool_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	now := time.Now()
	due := now.Add(24 * time.Hour)
	mock.ExpectQuery(`SELECT .* FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows(taskCols()).
			AddRow(uuid.New(), 1, "Task A", nil, "low", "normal", "1", "false", 2, 30, 0, due, now, now))

	items, err := r.ListByUserIDAndBoardPool(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1, got %d", len(items))
	}
	if items[0].BoardPool != 2 {
		t.Fatalf("expected BoardPool=2, got %d", items[0].BoardPool)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskListByUserIDAndBoardPool_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows(taskCols()))

	items, err := r.ListByUserIDAndBoardPool(context.Background(), 1, 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0, got %d", len(items))
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestTaskGetByID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	id := uuid.New()
	now := time.Now()
	due := now.Add(24 * time.Hour)
	mock.ExpectQuery(`SELECT .* FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows(taskCols()).
			AddRow(id, 1, "Task A", nil, "low", "normal", "1", "false", 0, 30, 0, due, now, now))

	task, err := r.GetByID(context.Background(), 1, id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.ID != id {
		t.Fatalf("expected ID %v, got %v", id, task.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTaskGetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "tasks"`).
		WillReturnRows(sqlmock.NewRows(taskCols()))

	_, err := r.GetByID(context.Background(), 1, uuid.New())
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestTaskDelete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "tasks"`).
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

func TestTaskDelete_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "tasks"`).
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

func TestTaskDelete_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewTaskRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "tasks"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	_, err := r.Delete(context.Background(), 1, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
