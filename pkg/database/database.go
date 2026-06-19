package database

import (
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Config carries the parameters needed to open a database connection.
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        string
}

// Open returns a GORM DB for the configured driver and DSN.
func Open(cfg Config, log *slog.Logger) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: newGORMLogger(cfg.LogLevel, log),
	}

	var db *gorm.DB
	var err error

	switch cfg.Driver {
	case "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(cfg.DSN), gormCfg)
	case "mysql", "mariadb":
		db, err = gorm.Open(mysql.Open(cfg.DSN), gormCfg)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DSN), gormCfg)
	case "sqlserver", "mssql":
		db, err = gorm.Open(sqlserver.Open(cfg.DSN), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported driver %q", cfg.Driver)
	}
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func newGORMLogger(level string, _ *slog.Logger) gormlogger.Interface {
	var logLevel gormlogger.LogLevel
	switch level {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	default:
		logLevel = gormlogger.Info
	}
	return gormlogger.Default.LogMode(logLevel)
}
