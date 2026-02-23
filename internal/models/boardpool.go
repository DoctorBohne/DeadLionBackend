package models

import (
	"time"

	"github.com/google/uuid"
)

type BoardPool struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	BoardID   uuid.UUID `gorm:"type:uuid;not null"`
	Title     string    `gorm:"type:text;not null"`
	Color     string    `gorm:"type:text;default:'000000'"`
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
