package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	repo "github.com/DoctorBohne/DeadLionBackend/internal/repositories/user"
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

func userCols() []string {
	return []string{
		"id", "created_at", "updated_at", "deleted_at",
		"sub", "issuer", "email_verified", "name",
		"preferred_username", "given_name", "family_name",
		"email", "onboarding_completed",
	}
}

// ── FindByIssuerSub ───────────────────────────────────────────────────────────

func TestFindByIssuerSub_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	now := time.Now()
	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows(userCols()).
			AddRow(1, now, now, nil, "sub1", "https://iss", false, "Alice", "alice", "Alice", "Smith", "a@b.com", false))

	u, err := r.FindByIssuerSub(context.Background(), "https://iss", "sub1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Sub != "sub1" {
		t.Fatalf("expected sub1, got %s", u.Sub)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIssuerSub_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows(userCols()))

	_, err := r.FindByIssuerSub(context.Background(), "https://iss", "nosub")
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIssuerSub_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnError(errors.New("db down"))

	_, err := r.FindByIssuerSub(context.Background(), "https://iss", "sub1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	u := &models.User{Sub: "sub1", Issuer: "https://iss", Email: "a@b.com"}
	if err := r.Create(context.Background(), u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreate_DBError(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	u := &models.User{Sub: "sub1", Issuer: "https://iss"}
	if err := r.Create(context.Background(), u); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── MarkOnboardingCompleted ───────────────────────────────────────────────────

func TestMarkOnboardingCompleted_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := r.MarkOnboardingCompleted(context.Background(), "https://iss", "sub1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMarkOnboardingCompleted_AlreadyCompleted(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	now := time.Now()
	// UPDATE affects 0 rows (onboarding_completed was already true)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	// follow-up SELECT to determine reason
	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows(userCols()).
			AddRow(1, now, now, nil, "sub1", "https://iss", false, "", "", "", "", "", true))

	err := r.MarkOnboardingCompleted(context.Background(), "https://iss", "sub1")
	if err == nil || err.Error() != "onboarding already completed" {
		t.Fatalf("expected 'onboarding already completed', got %v", err)
	}
	if err2 := mock.ExpectationsWereMet(); err2 != nil {
		t.Fatalf("unmet expectations: %v", err2)
	}
}

func TestMarkOnboardingCompleted_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows(userCols()))

	err := r.MarkOnboardingCompleted(context.Background(), "https://iss", "nosub")
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// ── UpdateEmailByIssuerSub ────────────────────────────────────────────────────

func TestUpdateEmailByIssuerSub_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := r.UpdateEmailByIssuerSub(context.Background(), "https://iss", "sub1", "new@example.com", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateEmailByIssuerSub_EmptyEmail(t *testing.T) {
	db, _ := newTestDB(t)
	r := repo.NewUserRepo(db)

	err := r.UpdateEmailByIssuerSub(context.Background(), "https://iss", "sub1", "   ", nil)
	if err == nil {
		t.Fatal("expected error for empty/whitespace email")
	}
}

func TestUpdateEmailByIssuerSub_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.UpdateEmailByIssuerSub(context.Background(), "https://iss", "nosub", "new@example.com", nil)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// ── SetEmailVerifiedByIssuerSub ───────────────────────────────────────────────

func TestSetEmailVerifiedByIssuerSub_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := r.SetEmailVerifiedByIssuerSub(context.Background(), "https://iss", "sub1", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSetEmailVerifiedByIssuerSub_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := r.SetEmailVerifiedByIssuerSub(context.Background(), "https://iss", "nosub", true)
	if !errors.Is(err, custom_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// ── FindOrCreateByIssuerSub ───────────────────────────────────────────────────

func TestFindOrCreateByIssuerSub_CreatesNew(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectCommit()

	u := &models.User{Sub: "newSub", Issuer: "https://iss", Email: "new@example.com"}
	result, created, err := r.FindOrCreateByIssuerSub(context.Background(), u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected created=true")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindOrCreateByIssuerSub_ReturnsExisting(t *testing.T) {
	db, mock := newTestDB(t)
	r := repo.NewUserRepo(db)

	now := time.Now()
	// INSERT ON CONFLICT DO NOTHING returns 0 rows
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()
	// follow-up SELECT for existing record
	mock.ExpectQuery(`SELECT .* FROM "users"`).
		WillReturnRows(sqlmock.NewRows(userCols()).
			AddRow(5, now, now, nil, "existing", "https://iss", false, "Ex", "ex", "E", "X", "ex@y.com", false))

	u := &models.User{Sub: "existing", Issuer: "https://iss"}
	result, created, err := r.FindOrCreateByIssuerSub(context.Background(), u)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Fatal("expected created=false")
	}
	if result.ID != 5 {
		t.Fatalf("expected ID=5, got %d", result.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
