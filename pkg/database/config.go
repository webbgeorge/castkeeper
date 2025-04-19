package database

import (
	"log/slog"

	slogGorm "github.com/orandin/slog-gorm"
	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConfigureDatabase(cfg config.Config, logger *slog.Logger) (*gorm.DB, error) {
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(logger.Handler()),
	)

	db, err := gorm.Open(
		dbDialector(cfg),
		&gorm.Config{
			TranslateError: true,
			Logger:         gormLogger,
		},
	)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&podcasts.Podcast{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&podcasts.Episode{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&auth.User{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&auth.Session{}); err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&framework.QueueTask{}); err != nil {
		return nil, err
	}

	return db, nil
}

func dbDialector(cfg config.Config) gorm.Dialector {
	switch cfg.Database.Driver {
	case config.DatabaseDriverPostgres:
		return postgres.New(postgres.Config{
			DSN:                  cfg.Database.DSN,
			PreferSimpleProtocol: true,
		})
	case config.DatabaseDriverSqlite:
		return sqlite.Open(cfg.Database.DSN)
	default:
		return nil
	}
}
