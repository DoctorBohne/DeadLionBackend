package abgabe_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	repo "github.com/DoctorBohne/DeadLionBackend/internal/repositories/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return gormDB, mock
}

func TestAbgabeRepo_Create_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "abgabes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	ab := &abgabe.Abgabe{
		Title:          "Math",
		DueDate:        time.Now().Add(24 * time.Hour),
		RiskAssessment: abgabe.Risk(2),
		UserID:         1,
	}
	if err := r.Create(context.Background(), ab); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAbgabeRepo_Create_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "abgabes"`)).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	ab := &abgabe.Abgabe{Title: "Math", UserID: 1}
	if err := r.Create(context.Background(), ab); err == nil {
		t.Fatal("expected error")
	}
}

func TestAbgabeRepo_FindByID_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	cols := []string{"id", "created_at", "updated_at", "deleted_at", "title", "due_date", "risk_assessment", "user_id"}
	mock.ExpectQuery(`SELECT .* FROM "abgabes"`).
		WithArgs(uint(5), 1).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(5, time.Now(), time.Now(), nil, "Math", time.Now(), 2, 1))

	result, err := r.FindByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != 1 {
		t.Fatalf("unexpected userID: %d", result.UserID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAbgabeRepo_FindByID_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "abgabes"`).
		WithArgs(uint(99), 1).
		WillReturnRows(sqlmock.NewRows(nil))

	_, err := r.FindByID(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAbgabeRepo_ListByUser_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	cols := []string{"id", "created_at", "updated_at", "deleted_at", "title", "due_date", "risk_assessment", "user_id"}
	mock.ExpectQuery(`SELECT .* FROM "abgabes"`).
		WithArgs(uint(1)).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(1, time.Now(), time.Now(), nil, "Math", time.Now(), 2, 1).
			AddRow(2, time.Now(), time.Now(), nil, "Phys", time.Now(), 1, 1))

	items, err := r.ListByUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestAbgabeRepo_ListByUser_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "abgabes"`).
		WillReturnError(gorm.ErrInvalidDB)

	_, err := r.ListByUser(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAbgabeRepo_Update_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "abgabes"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	ab := &abgabe.Abgabe{Model: gorm.Model{ID: 1}, Title: "Updated", UserID: 1}
	if err := r.Update(context.Background(), ab); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAbgabeRepo_Delete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "abgabes" SET "deleted_at"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	ab := &abgabe.Abgabe{Model: gorm.Model{ID: 1}, Title: "Math", UserID: 1}
	if err := r.Delete(context.Background(), ab); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAbgabeRepo_ListByUserAndDateBeforeFromNow_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewAbgabeRepo(db)

	cols := []string{"id", "created_at", "updated_at", "deleted_at", "title", "due_date", "risk_assessment", "user_id"}
	now := time.Now()
	future := now.Add(30 * 24 * time.Hour)

	mock.ExpectQuery(`SELECT .* FROM "abgabes"`).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(1, now, now, nil, "Math", future, 2, 1))

	items, err := r.ListByUserAndDateBeforeFromNow(context.Background(), 1, now, future)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}
