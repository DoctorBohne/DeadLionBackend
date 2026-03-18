package db

import (
	"github.com/DoctorBohne/DeadLionBackend/internal/abgabe"
	"github.com/DoctorBohne/DeadLionBackend/internal/models"
	"gorm.io/gorm"
)

func Migrate(gdb *gorm.DB) error {
	if err := gdb.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
		return err
	}

	if err := gdb.AutoMigrate(
		&models.User{},
		&abgabe.Abgabe{},
		&abgabe.UniversityModule{},
		&models.Userboard{},
		&models.BoardPool{},
		&models.Task{},
		&models.Subtask{},
		&models.Taskboard{},
	); err != nil {
		return err
	}
	return nil
}
