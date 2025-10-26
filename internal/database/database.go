package database

import (
	"fmt"
	"time"

	"developer-portal-backend/internal/database/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Options struct {
	LogLevel        logger.LogLevel
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	AutoMigrate     bool
}

// Initialize opens a Postgres connection and creates the schema from GORM models.
// Simplified single-phase AutoMigrate since cyclic foreign keys were removed.
func Initialize(dsn string, opts *Options) (*gorm.DB, error) {
	// Defaults
	if opts == nil {
		opts = &Options{}
	}
	if opts.LogLevel == 0 {
		opts.LogLevel = logger.Error
	}
	if opts.MaxOpenConns == 0 {
		opts.MaxOpenConns = 20
	}
	if opts.MaxIdleConns == 0 {
		opts.MaxIdleConns = 10
	}
	if opts.ConnMaxLifetime == 0 {
		opts.ConnMaxLifetime = 30 * time.Minute
	}
	if opts.ConnMaxIdleTime == 0 {
		opts.ConnMaxIdleTime = 10 * time.Minute
	}
	if !opts.AutoMigrate {
		opts.AutoMigrate = true
	}

	// Open DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(opts.LogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
		sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(opts.ConnMaxIdleTime)
	}

	// Ensure required extension for UUID generation (used by BaseModel default gen_random_uuid())
	_ = db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error

	// AutoMigrate all models (no cycles)
	if opts.AutoMigrate {
		all := []interface{}{
			&models.Organization{},
			&models.Group{},
			&models.Member{},
			&models.Team{},
			&models.Landscape{},
			&models.Project{},
			&models.Component{},
			&models.ProjectLandscape{},
			&models.ProjectComponent{},
			&models.TeamComponentOwnership{},
			&models.TeamLeadership{},
			&models.ComponentDeployment{},
			&models.DeploymentTimeline{},
			&models.DutySchedule{},
			&models.OutageCall{},
			&models.OutageCallAssignee{},
		}
		if err := db.AutoMigrate(all...); err != nil {
			return nil, fmt.Errorf("auto-migrate: %w", err)
		}
	}

	return db, nil
}
