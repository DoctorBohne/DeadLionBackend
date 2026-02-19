package models

import (
	"time"

	"github.com/google/uuid"
)

type Taskboard struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v2()"`
	TaskID      uuid.UUID `gorm:"type:uuid;not null"`
	Title       string    `gorm:"type:text;not null"`
	Description *string   `gorm:"type:text"`
	Status      string    `gorm:"type:varchar(50);default:'todo'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
