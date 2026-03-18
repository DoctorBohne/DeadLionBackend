package abgabe

import (
	"time"

	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"gorm.io/gorm"
)

type Risk int

const (
	SehrNiedrig Risk = iota
	Niedrig
	Mittel
	Hoch
	SehrHoch
)

// Abgabe Model this contains all information about the regular "tasks"
// belongs to one user only and one user can have many tasks
type Abgabe struct {
	gorm.Model
	Title          string      `json:"title"`
	DueDate        time.Time   `json:"due_date"`
	RiskAssessment Risk        `json:"risk_assessment"`
	UserID         uint        `gorm:"not null;index"`
	User           models.User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
