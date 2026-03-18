package models

import (
	"time"

	"github.com/google/uuid"
)

type Subtask struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskID      uuid.UUID `gorm:"type:uuid;not null"`
	Title       string    `gorm:"type:text;not null"`
	Description *string   `gorm:"type:text"`
	BoardPool   int       `gorm:"type:int;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
