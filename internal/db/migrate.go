package db

import (
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"gorm.io/gorm"
)

func Migrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&models.User{},
	)
}
