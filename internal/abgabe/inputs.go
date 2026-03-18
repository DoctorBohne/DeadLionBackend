package abgabe

import (
	"time"
)

type CreateAbgabeInput struct {
	Title          string    `json:"title" binding:"required"`
	DueDate        time.Time `json:"due_date" binding:"required"`
	RiskAssessment int       `json:"risk_assessment" binding:"required"`
}

type UpdateAbgabeInput struct {
	Title          *string    `json:"title"`
	DueDate        *time.Time `json:"due_date"`
	RiskAssessment *int       `json:"risk_assessment"`
}
