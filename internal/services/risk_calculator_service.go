package services

import (
	"context"
	"errors"
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
)

type RiskCalculatorService struct {
	repo AbgabeRepo
	// exam repo doesn't exist yet, if, include here
}

func NewRiskCalculatorService(repo AbgabeRepo) *RiskCalculatorService {
	return &RiskCalculatorService{repo: repo}
}

func (s *RiskCalculatorService) CalculateRiskList(ctx context.Context, userID uint, requestDate time.Time) ([]models.RiskItem, error) {

	var list []abgabe.Abgabe
	var err error
	if requestDate.IsZero() {
		list, err = s.repo.ListByUser(ctx, userID)
		if err != nil {
			return nil, err
		}
	} else {
		list, err = s.repo.ListByUserAndDateBeforeFromNow(ctx, userID, time.Now(), requestDate)
		if err != nil {
			return nil, err
		}
	}
	return s.convertSubmissionListToRiskItemList(list), nil

	//TODO: add Exams when they exist
}

func (s *RiskCalculatorService) convertSubmissionListToRiskItemList(abgabeList []abgabe.Abgabe) []models.RiskItem {
	riskItemList := make([]models.RiskItem, 0, len(abgabeList))
	for i := range abgabeList {
		riskItem := s.convertSubmissionEntryToRiskItem(&abgabeList[i])
		err := s.calculateRiskScore(riskItem)
		if err != nil {
			continue
		}
		riskItemList = append(riskItemList, *riskItem)
	}
	return riskItemList
}

func (s *RiskCalculatorService) convertSubmissionEntryToRiskItem(submissionEntry *abgabe.Abgabe) *models.RiskItem {
	return &models.RiskItem{
		ID:       submissionEntry.ID,
		Type:     "abgabe",
		Title:    submissionEntry.Title,
		Deadline: submissionEntry.DueDate,
		Priority: int(submissionEntry.RiskAssessment),
		ModuleID: submissionEntry.ModulID,
	}
}

func (s *RiskCalculatorService) calculateRiskScore(item *models.RiskItem) error {
	urgency, err := s.extractUrgencyScore(item.Deadline)
	if err != nil {
		return err
	}
	priority := item.Priority

	item.RiskScore = float64((2 * priority) + (3 * urgency))
	return nil

}

func (s *RiskCalculatorService) extractUrgencyScore(date time.Time) (int, error) {
	if date.IsZero() {
		return 0, errors.New("no Date provided")
	}
	duration := date.Sub(time.Now())
	days := int(duration.Hours() / 24)

	switch {
	case days > 21:
		return 1, nil
	case days > 14:
		return 2, nil
	case days > 7:
		return 3, nil
	case days > 2:
		return 4, nil
	case days > 0:
		return 5, nil
	default:
		return 0, errors.New("invalid date")
	}
}
