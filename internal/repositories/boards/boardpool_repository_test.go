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

func boardPoolCols() []string {
	return []string{"id", "board_id", "title", "color", "position", "created_at", "updated_at"}
}

func newBoardPool(boardID uuid.UUID) *models.BoardPool {
	return &models.BoardPool{
		ID:        uuid.New(),
		BoardID:   boardID,
		Title:     "To Do",
		Color:     "ff0000",
		Position:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ── Create (append, explicitPosition=false) ───────────────────────────────────

func TestBoardPoolCreate_Append_Success(t *testing.T) {
	// Use DisableNestedTransaction so tx.Create() inside Transaction() does not
	// issue a SAVEPOINT, keeping the mock sequence simple.
	db, mock := newTestDBNoNestedTx(t)
	r := boards.NewBoardPoolRepo(db)

	boardID := uuid.New()
	newID := uuid.New()

	// ownership check
	mock.ExpectQuery(`SELECT count\(\*\) FROM "userboards"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// outer transaction
	mock.ExpectBegin()
	// SELECT MAX(position)
	mock.ExpectQuery(`SELECT MAX\(position\)`).
		WillReturnRows(sqlmock.NewRows([]string{"MAX(position)"}).AddRow(2))
	// INSERT board_pool
	mock.ExpectQuery(`INSERT INTO "board_pools"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(newID))
	mock.ExpectCommit()

	p := newBoardPool(boardID)
	if err := r.Create(context.Background(), 1, p, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Position != 3 {
		t.Fatalf("expected position=3 (max+1), got %d", p.Position)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBoardPoolCreate_ExplicitPosition_Success(t *testing.T) {
	db, mock := newTestDBNoNestedTx(t)
	r := boards.NewBoardPoolRepo(db)

	boardID := uuid.New()
	newID := uuid.New()

	// ownership check
	mock.ExpectQuery(`SELECT count\(\*\) FROM "userboards"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// outer transaction
	mock.ExpectBegin()
	// shift existing lanes
	mock.ExpectExec(`UPDATE "board_pools"`).
		WillReturnResult(sqlmock.NewResult(2, 2))
	// INSERT board_pool
	mock.ExpectQuery(`INSERT INTO "board_pools"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(newID))
	mock.ExpectCommit()

	p := newBoardPool(boardID)
	p.Position = 1
	if err := r.Create(context.Background(), 1, p, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBoardPoolCreate_BoardNotOwned(t *testing.T) {
	db, mock := newTestDBNoNestedTx(t)
	r := boards.NewBoardPoolRepo(db)

	// COUNT returns 0 — board doesn't belong to user
	mock.ExpectQuery(`SELECT count\(\*\) FROM "userboards"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	p := newBoardPool(uuid.New())
	if err := r.Create(context.Background(), 1, p, false); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── ListByBoardID ─────────────────────────────────────────────────────────────

func TestBoardPoolListByBoardID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	boardID := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "board_pools"`).
		WillReturnRows(sqlmock.NewRows(boardPoolCols()).
			AddRow(uuid.New(), boardID, "Todo", "000000", 0, now, now).
			AddRow(uuid.New(), boardID, "Done", "00ff00", 1, now, now))

	items, err := r.ListByBoardID(context.Background(), 1, boardID)
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

func TestBoardPoolListByBoardID_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "board_pools"`).
		WillReturnRows(sqlmock.NewRows(boardPoolCols()))

	items, err := r.ListByBoardID(context.Background(), 1, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestBoardPoolListByBoardID_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "board_pools"`).
		WillReturnError(errors.New("db down"))

	_, err := r.ListByBoardID(context.Background(), 1, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestBoardPoolGetByID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	id := uuid.New()
	boardID := uuid.New()
	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "board_pools"`).
		WillReturnRows(sqlmock.NewRows(boardPoolCols()).
			AddRow(id, boardID, "Todo", "000000", 0, now, now))

	p, err := r.GetByID(context.Background(), 1, boardID, id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != id {
		t.Fatalf("expected ID %v, got %v", id, p.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBoardPoolGetByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "board_pools"`).
		WillReturnRows(sqlmock.NewRows(boardPoolCols()))

	_, err := r.GetByID(context.Background(), 1, uuid.New(), uuid.New())
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestBoardPoolUpdate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	id := uuid.New()
	boardID := uuid.New()
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "board_pools"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT .* FROM "board_pools"`).
		WillReturnRows(sqlmock.NewRows(boardPoolCols()).
			AddRow(id, boardID, "Renamed", "ffffff", 0, now, now))

	p, err := r.Update(context.Background(), 1, boardID, id, map[string]any{"title": "Renamed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Title != "Renamed" {
		t.Fatalf("expected Renamed, got %s", p.Title)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBoardPoolUpdate_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "board_pools"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	_, err := r.Update(context.Background(), 1, uuid.New(), uuid.New(), map[string]any{"title": "X"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestBoardPoolDelete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "board_pools"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	deleted, err := r.Delete(context.Background(), 1, uuid.New(), uuid.New())
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

func TestBoardPoolDelete_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "board_pools"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	deleted, err := r.Delete(context.Background(), 1, uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted {
		t.Fatal("expected deleted=false (not found)")
	}
}

func TestBoardPoolDelete_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := boards.NewBoardPoolRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "board_pools"`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	_, err := r.Delete(context.Background(), 1, uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
