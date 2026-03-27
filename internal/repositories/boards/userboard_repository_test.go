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

func userboardCols() []string {
	return []string{"id", "user_id", "title", "description", "created_at", "updated_at"}
}

func newUserboard(userID uint) *models.Userboard {
	desc := "a board"
	return &models.Userboard{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       "Test Board",
		Description: &desc,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestUserboardCreate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	id := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "userboards"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectCommit()

	b := newUserboard(1)
	if err := r.Create(context.Background(), b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserboardCreate_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "userboards"`).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	b := newUserboard(1)
	if err := r.Create(context.Background(), b); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── ListByUserID ──────────────────────────────────────────────────────────────

func TestUserboardListByUserID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "userboards"`).
		WillReturnRows(sqlmock.NewRows(userboardCols()).
			AddRow(uuid.New(), 1, "Board A", nil, now, now).
			AddRow(uuid.New(), 1, "Board B", nil, now, now))

	items, err := r.ListByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserboardListByUserID_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "userboards"`).
		WillReturnRows(sqlmock.NewRows(userboardCols()))

	items, err := r.ListByUserID(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestUserboardListByUserID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "userboards"`).
		WillReturnError(errors.New("db down"))

	_, err := r.ListByUserID(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestUserboardGetByID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	id := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "userboards"`).
		WillReturnRows(sqlmock.NewRows(userboardCols()).
			AddRow(id, 1, "Board A", nil, now, now))

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

func TestUserboardGetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "userboards"`).
		WillReturnRows(sqlmock.NewRows(userboardCols()))

	_, err := r.GetByID(context.Background(), 1, uuid.New())
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUserboardUpdate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	id := uuid.New()
	now := time.Now()
	// UPDATE
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "userboards"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// follow-up GetByID
	mock.ExpectQuery(`SELECT .* FROM "userboards"`).
		WillReturnRows(sqlmock.NewRows(userboardCols()).
			AddRow(id, 1, "Updated", nil, now, now))

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

func TestUserboardUpdate_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "userboards"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	_, err := r.Update(context.Background(), 1, uuid.New(), map[string]any{"title": "X"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestUserboardDelete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "userboards"`).
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

func TestUserboardDelete_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "userboards"`).
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

func TestUserboardDelete_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewUserboardRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "userboards"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	_, err := r.Delete(context.Background(), 1, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
