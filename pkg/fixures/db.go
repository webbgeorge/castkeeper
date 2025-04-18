package fixures

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path"

	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

func ConfigureDBForTestWithFixtures() (db *gorm.DB, resetFn func()) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	dbPath := path.Join(fixtureDir(), "../../data/test-unit.db")
	db, err := database.ConfigureDatabase(config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    dbPath,
		},
	}, logger)
	if err != nil {
		panic(err)
	}

	applyFixtures(db)

	return db, func() {
		_ = os.Remove(dbPath)
	}
}

func applyFixtures(db *gorm.DB) {
	// use the testdata feed service to get fixture podcast for convenience
	feedService := podcasts.FeedService{
		HTTPClient: TestDataHTTPClient,
	}
	pod, eps, err := feedService.ParseFeed(context.Background(), "http://testdata/feeds/valid.xml")
	if err != nil {
		panic(err)
	}

	create(db, &pod)
	for _, ep := range eps {
		ep.Status = podcasts.EpisodeStatusSuccess
		create(db, &ep)
	}
}

func create(db *gorm.DB, value any) {
	if err := db.Create(value).Error; err != nil {
		panic(err)
	}
}
