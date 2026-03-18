package db

import (
	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"gorm.io/gorm"
)

func Migrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&models.User{},
		&abgabe.Abgabe{},
		&abgabe.UniversityModule{},
		&models.Userboard{},
		&models.BoardPool{},
		&models.Task{},
		&models.Subtask{},
		&models.Taskboard{},
	)
}
