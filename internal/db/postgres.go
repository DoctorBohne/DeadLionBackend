package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresConfig struct {
	Host    string
	Port    string
	User    string
	Pass    string
	Name    string
	SSLMode string

	MaxOpenConns int
	MaxIdleConns int
	ConnMaxIdle  time.Duration
	ConnMaxLife  time.Duration
	LogLevel     logger.LogLevel
	TimeZone     string
}

func OpenPostgres(cfg PostgresConfig) (*gorm.DB, error) {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxIdle == 0 {
		cfg.ConnMaxIdle = 5 * time.Minute
	}
	if cfg.ConnMaxLife == 0 {
		cfg.ConnMaxLife = 30 * time.Minute
	}
	if cfg.LogLevel == 0 {
		cfg.LogLevel = logger.Warn
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = "Europe/Berlin"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.Name, cfg.SSLMode, cfg.TimeZone,
	)

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(cfg.LogLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdle)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLife)

	if err = sqlDB.Ping(); err != nil {
		return nil, err
	}

	return gdb, nil
}
