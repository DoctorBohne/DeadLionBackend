package models

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v2()"`
	UserID          uint
	Title           string  `gorm:"type:text;not null"`
	Description     *string `gorm:"type:text"`
	Risk            string  `gorm:"type:varchar(50);default:'todo'"`
	Priority        string
	PriorityRank    string
	IsFinished      string    `gorm:"type:text;not null"`
	BoardPool       int       `gorm:"type:int;not null"`
	EstimateMinutes int       `gorm:"type:int;not null"`
	SpendMinutes    int       `gorm:"type:int;not null"`
	DueAt           time.Time `gorm:"type:datetime;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
