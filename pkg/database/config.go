package database

import (
	"log/slog"
	"os"
	"path"

	slogGorm "github.com/orandin/slog-gorm"
	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConfigureDatabase(cfg config.Config, logger *slog.Logger, inMemory bool) (*gorm.DB, error) {
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(logger.Handler()),
	)

	dialector, err := dbDialector(cfg, inMemory)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(
		dialector,
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
	if err := db.AutoMigrate(&framework.ScheduledTaskState{}); err != nil {
		return nil, err
	}

	return db, nil
}

func dbDialector(cfg config.Config, inMemory bool) (gorm.Dialector, error) {
	if inMemory {
		return sqlite.Open(":memory:"), nil
	}

	err := os.MkdirAll(cfg.DataDirPath, 0750)
	if err != nil {
		return nil, err
	}
	dsn := path.Join(cfg.DataDirPath, "data.db")
	return sqlite.Open(dsn), nil
}
