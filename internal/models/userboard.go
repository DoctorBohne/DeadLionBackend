package models

import (
	"time"

	"github.com/google/uuid"
)

type Userboard struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v2()"`
	UserID      uint      `gorm:"not null;index"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Title       string    `gorm:"type:text;not null"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
