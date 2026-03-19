package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"gorm.io/gorm"
)

// ---- RiskCalculatorService tests ----

func TestRiskCalculatorService_CalculateRiskList_ZeroDate(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour)
	repo := &stubAbgabeRepo{
		listByUserFn: func(_ context.Context, userID uint) ([]abgabe.Abgabe, error) {
			return []abgabe.Abgabe{
				{Model: gorm.Model{ID: 1}, UserID: userID, Title: "A", DueDate: future, RiskAssessment: abgabe.Risk(2)},
			}, nil
		},
	}
	svc := NewRiskCalculatorService(repo)
	list, err := svc.CalculateRiskList(context.Background(), 1, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 risk item, got %d", len(list))
	}
}

func TestRiskCalculatorService_CalculateRiskList_WithDate(t *testing.T) {
	future := time.Now().Add(30 * 24 * time.Hour)
	repo := &stubAbgabeRepo{
		listByUserAndDateBeforeFromNowFn: func(_ context.Context, userID uint, _, _ time.Time) ([]abgabe.Abgabe, error) {
			return []abgabe.Abgabe{
				{Model: gorm.Model{ID: 2}, UserID: userID, Title: "B", DueDate: future, RiskAssessment: abgabe.Risk(3)},
			}, nil
		},
	}
	svc := NewRiskCalculatorService(repo)
	requestDate := time.Now().Add(60 * 24 * time.Hour)
	list, err := svc.CalculateRiskList(context.Background(), 1, requestDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 risk item, got %d", len(list))
	}
}

func TestRiskCalculatorService_CalculateRiskList_RepoError(t *testing.T) {
	repo := &stubAbgabeRepo{
		listByUserFn: func(_ context.Context, _ uint) ([]abgabe.Abgabe, error) {
			return nil, errors.New("db error")
		},
	}
	svc := NewRiskCalculatorService(repo)
	_, err := svc.CalculateRiskList(context.Background(), 1, time.Time{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRiskCalculatorService_CalculateRiskList_SkipsZeroDeadline(t *testing.T) {
	// items with zero deadline fail extractUrgencyScore and are skipped
	repo := &stubAbgabeRepo{
		listByUserFn: func(_ context.Context, userID uint) ([]abgabe.Abgabe, error) {
			return []abgabe.Abgabe{
				{Model: gorm.Model{ID: 1}, UserID: userID, Title: "Expired", DueDate: time.Time{}, RiskAssessment: abgabe.Risk(1)},
			}, nil
		},
	}
	svc := NewRiskCalculatorService(repo)
	list, err := svc.CalculateRiskList(context.Background(), 1, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 items, got %d", len(list))
	}
}

func TestRiskCalculatorService_CalculateRiskList_RiskScoreCalculated(t *testing.T) {
	// 30 days ahead: urgency=1, priority=3 → score = 2*3 + 3*1 = 9
	future := time.Now().Add(30 * 24 * time.Hour)
	repo := &stubAbgabeRepo{
		listByUserFn: func(_ context.Context, userID uint) ([]abgabe.Abgabe, error) {
			return []abgabe.Abgabe{
				{Model: gorm.Model{ID: 1}, UserID: userID, Title: "A", DueDate: future, RiskAssessment: abgabe.Risk(3)},
			}, nil
		},
	}
	svc := NewRiskCalculatorService(repo)
	list, err := svc.CalculateRiskList(context.Background(), 1, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list))
	}
	expected := float64(2*3 + 3*1)
	if list[0].RiskScore != expected {
		t.Fatalf("expected risk score %v, got %v", expected, list[0].RiskScore)
	}
}
