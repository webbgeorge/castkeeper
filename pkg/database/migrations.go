package database

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/database/migrations"
	"gorm.io/gorm"
)

type migration interface {
	Name() string
	Migrate(db *gorm.DB) error
}

// migrations are run in this order
var allMigrations = []migration{
	migrations.Migration001Init{},
}

type appliedMigration struct {
	Name      string `gorm:"primarykey"`
	AppliedAt time.Time
}

func migrate(db *gorm.DB, logger *slog.Logger) error {
	// auto-migrate the applied_migrations table only
	if err := db.AutoMigrate(&appliedMigration{}); err != nil {
		return err
	}

	appliedMigrations, err := listAppliedMigrations(db)
	if err != nil {
		return err
	}

	for _, m := range allMigrations {
		if slices.ContainsFunc(appliedMigrations, func(am appliedMigration) bool {
			return am.Name == m.Name()
		}) {
			continue
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Migrate(tx); err != nil {
				return err
			}

			if err := tx.Create(&appliedMigration{
				Name:      m.Name(),
				AppliedAt: time.Now(),
			}).Error; err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			logger.Error(fmt.Sprintf(
				"Failed to apply migration '%s': %s",
				m.Name(),
				err.Error(),
			))
			return err
		}
		logger.Info(fmt.Sprintf(
			"Successfully applied migration '%s'",
			m.Name(),
		))
	}

	return nil
}

func listAppliedMigrations(db *gorm.DB) ([]appliedMigration, error) {
	var appliedMigrations []appliedMigration
	result := db.
		Order("name asc").
		Find(&appliedMigrations)
	if result.Error != nil {
		return nil, result.Error
	}
	return appliedMigrations, nil
}
