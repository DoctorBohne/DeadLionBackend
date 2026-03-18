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
	if gdb.Migrator().HasColumn(&abgabe.Abgabe{}, "modul_id") {
		if err := gdb.Migrator().DropColumn(&abgabe.Abgabe{}, "modul_id"); err != nil {
			return err
		}
	}
	if gdb.Migrator().HasTable("university_modules") {
		if err := gdb.Migrator().DropTable("university_modules"); err != nil {
			return err
		}
	}
	return nil
}
