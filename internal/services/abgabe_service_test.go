package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/custom_errors"
	"gorm.io/gorm"
)

// ---- stub AbgabeRepo ----

type stubAbgabeRepo struct {
	createFn                        func(ctx context.Context, ab *abgabe.Abgabe) error
	findByIDFn                      func(ctx context.Context, id uint) (*abgabe.Abgabe, error)
	listByUserFn                    func(ctx context.Context, userID uint) ([]abgabe.Abgabe, error)
	listByUserAndDateBeforeFn       func(ctx context.Context, userID uint, date time.Time) ([]abgabe.Abgabe, error)
	updateFn                        func(ctx context.Context, ab *abgabe.Abgabe) error
	listByUserAndDateBeforeFromNowFn func(ctx context.Context, userID uint, now, requestDate time.Time) ([]abgabe.Abgabe, error)
	deleteFn                        func(ctx context.Context, ab *abgabe.Abgabe) error
}

func (s *stubAbgabeRepo) Create(ctx context.Context, ab *abgabe.Abgabe) error {
	return s.createFn(ctx, ab)
}
func (s *stubAbgabeRepo) FindByID(ctx context.Context, id uint) (*abgabe.Abgabe, error) {
	return s.findByIDFn(ctx, id)
}
func (s *stubAbgabeRepo) ListByUser(ctx context.Context, userID uint) ([]abgabe.Abgabe, error) {
	return s.listByUserFn(ctx, userID)
}
func (s *stubAbgabeRepo) ListByUserAndDateBefore(ctx context.Context, userID uint, date time.Time) ([]abgabe.Abgabe, error) {
	return s.listByUserAndDateBeforeFn(ctx, userID, date)
}
func (s *stubAbgabeRepo) Update(ctx context.Context, ab *abgabe.Abgabe) error {
	return s.updateFn(ctx, ab)
}
func (s *stubAbgabeRepo) ListByUserAndDateBeforeFromNow(ctx context.Context, userID uint, now, requestDate time.Time) ([]abgabe.Abgabe, error) {
	return s.listByUserAndDateBeforeFromNowFn(ctx, userID, now, requestDate)
}
func (s *stubAbgabeRepo) Delete(ctx context.Context, ab *abgabe.Abgabe) error {
	return s.deleteFn(ctx, ab)
}

// ---- AbgabeService tests ----

func TestAbgabeService_Create_Success(t *testing.T) {
	repo := &stubAbgabeRepo{
		createFn: func(_ context.Context, ab *abgabe.Abgabe) error { return nil },
	}
	svc := AbgabeService{r: repo}

	result, err := svc.Create(context.Background(), 1, abgabe.CreateAbgabeInput{
		Title:          "Math",
		DueDate:        time.Now().Add(48 * time.Hour),
		RiskAssessment: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Math" || result.UserID != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestAbgabeService_Create_RepoError(t *testing.T) {
	repo := &stubAbgabeRepo{
		createFn: func(_ context.Context, _ *abgabe.Abgabe) error {
			return errors.New("db error")
		},
	}
	svc := AbgabeService{r: repo}
	_, err := svc.Create(context.Background(), 1, abgabe.CreateAbgabeInput{Title: "Math"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAbgabeService_Get_Success(t *testing.T) {
	repo := &stubAbgabeRepo{
		findByIDFn: func(_ context.Context, id uint) (*abgabe.Abgabe, error) {
			return &abgabe.Abgabe{Model: gorm.Model{ID: id}, UserID: 1, Title: "Math"}, nil
		},
	}
	svc := AbgabeService{r: repo}
	result, err := svc.Get(context.Background(), 1, 10)
	if err != nil || result.Title != "Math" {
		t.Fatalf("unexpected result: %v, %v", result, err)
	}
}

func TestAbgabeService_Get_Forbidden(t *testing.T) {
	repo := &stubAbgabeRepo{
		findByIDFn: func(_ context.Context, id uint) (*abgabe.Abgabe, error) {
			return &abgabe.Abgabe{Model: gorm.Model{ID: id}, UserID: 999, Title: "Other"}, nil
		},
	}
	svc := AbgabeService{r: repo}
	_, err := svc.Get(context.Background(), 1, 10)
	if !errors.Is(err, custom_errors.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAbgabeService_Get_RepoError(t *testing.T) {
	repo := &stubAbgabeRepo{
		findByIDFn: func(_ context.Context, _ uint) (*abgabe.Abgabe, error) {
			return nil, errors.New("not found")
		},
	}
	svc := AbgabeService{r: repo}
	_, err := svc.Get(context.Background(), 1, 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAbgabeService_List_Success(t *testing.T) {
	repo := &stubAbgabeRepo{
		listByUserFn: func(_ context.Context, userID uint) ([]abgabe.Abgabe, error) {
			return []abgabe.Abgabe{{UserID: userID, Title: "A"}, {UserID: userID, Title: "B"}}, nil
		},
	}
	svc := AbgabeService{r: repo}
	items, err := svc.List(context.Background(), 1)
	if err != nil || len(items) != 2 {
		t.Fatalf("unexpected: %v %v", items, err)
	}
}

func TestAbgabeService_List_RepoError(t *testing.T) {
	repo := &stubAbgabeRepo{
		listByUserFn: func(_ context.Context, _ uint) ([]abgabe.Abgabe, error) {
			return nil, errors.New("db error")
		},
	}
	svc := AbgabeService{r: repo}
	_, err := svc.List(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAbgabeService_Update_Success(t *testing.T) {
	newTitle := "Updated"
	repo := &stubAbgabeRepo{
		findByIDFn: func(_ context.Context, id uint) (*abgabe.Abgabe, error) {
			return &abgabe.Abgabe{Model: gorm.Model{ID: id}, UserID: 1, Title: "Old"}, nil
		},
		updateFn: func(_ context.Context, _ *abgabe.Abgabe) error { return nil },
	}
	svc := AbgabeService{r: repo}
	result, err := svc.Update(context.Background(), 1, 10, abgabe.UpdateAbgabeInput{Title: &newTitle})
	if err != nil || result.Title != "Updated" {
		t.Fatalf("unexpected: %v %v", result, err)
	}
}

func TestAbgabeService_Update_Forbidden(t *testing.T) {
	repo := &stubAbgabeRepo{
		findByIDFn: func(_ context.Context, id uint) (*abgabe.Abgabe, error) {
			return &abgabe.Abgabe{Model: gorm.Model{ID: id}, UserID: 999}, nil
		},
	}
	svc := AbgabeService{r: repo}
	_, err := svc.Update(context.Background(), 1, 10, abgabe.UpdateAbgabeInput{})
	if !errors.Is(err, custom_errors.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAbgabeService_Delete_Success(t *testing.T) {
	repo := &stubAbgabeRepo{
		findByIDFn: func(_ context.Context, id uint) (*abgabe.Abgabe, error) {
			return &abgabe.Abgabe{Model: gorm.Model{ID: id}, UserID: 1}, nil
		},
		deleteFn: func(_ context.Context, _ *abgabe.Abgabe) error { return nil },
	}
	svc := AbgabeService{r: repo}
	if err := svc.Delete(context.Background(), 1, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAbgabeService_Delete_Forbidden(t *testing.T) {
	repo := &stubAbgabeRepo{
		findByIDFn: func(_ context.Context, id uint) (*abgabe.Abgabe, error) {
			return &abgabe.Abgabe{Model: gorm.Model{ID: id}, UserID: 999}, nil
		},
	}
	svc := AbgabeService{r: repo}
	err := svc.Delete(context.Background(), 1, 10)
	if !errors.Is(err, custom_errors.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAbgabeService_Delete_RepoFindError(t *testing.T) {
	repo := &stubAbgabeRepo{
		findByIDFn: func(_ context.Context, _ uint) (*abgabe.Abgabe, error) {
			return nil, errors.New("not found")
		},
	}
	svc := AbgabeService{r: repo}
	if err := svc.Delete(context.Background(), 1, 10); err == nil {
		t.Fatal("expected error")
	}
}
