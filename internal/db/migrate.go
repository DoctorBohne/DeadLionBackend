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
		&models.Userboard{},
		&models.BoardPool{},
		&models.Task{},
		&models.Subtask{},
		&models.Taskboard{},
	); err != nil {
		return err
	}
	if err := gdb.Exec(`ALTER TABLE abgabes DROP COLUMN IF EXISTS modul_id CASCADE`).Error; err != nil {
		return err
	}
	if err := gdb.Exec(`DROP TABLE IF EXISTS university_modules CASCADE`).Error; err != nil {
		return err
	}
	return nil
}
