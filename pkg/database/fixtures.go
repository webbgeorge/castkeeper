package database

import (
	"os"

	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

const unitTestDBPath = "data/test-unit.db"

func ConfigureDBForTestWithFixtures() (db *gorm.DB, resetFn func()) {
	db, err := ConfigureDatabase(config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    unitTestDBPath,
		},
	}, nil)
	if err != nil {
		panic(err)
	}

	applyFixtures(db)

	return db, func() {
		os.Remove(unitTestDBPath)
	}
}

// TODO circular dependency. Will need to be moved to diff package
func applyFixtures(db *gorm.DB) {
	// TODO
	create(db, &podcasts.Podcast{
		// TODO
	})
}

func create(db *gorm.DB, value interface{}) {
	if err := db.Create(value).Error; err != nil {
		panic(err)
	}
}
