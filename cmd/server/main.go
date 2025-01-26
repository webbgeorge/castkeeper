package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	slogGorm "github.com/orandin/slog-gorm"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/downloadworker"
	"github.com/webbgeorge/castkeeper/pkg/feedworker"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/webserver"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, logger, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	db, err := configureDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("failed to connect to database", err)
	}

	objstore := &objectstorage.LocalObjectStorage{
		HTTPClient: framework.NewHTTPClient(time.Second * 5),
		BasePath:   cfg.ObjectStorage.LocalBasePath,
	}
	feedService := &podcasts.FeedService{
		HTTPClient: framework.NewHTTPClient(time.Second * 5),
	}
	itunesAPI := &itunes.ItunesAPI{
		HTTPClient: framework.NewHTTPClient(time.Second * 5),
	}

	ctx := framework.ContextWithLogger(context.Background(), logger)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return webserver.Start(ctx, logger, feedService, db, objstore, itunesAPI)
	})

	g.Go(func() error {
		fw := feedworker.FeedWorker{
			FeedService: feedService,
			DB:          db,
		}
		return fw.Start(ctx)
	})

	g.Go(func() error {
		// dw := downloadworker.DownloadWorker{
		_ = downloadworker.DownloadWorker{
			DB: db,
			OS: objstore,
		}
		// return dw.Start(ctx)
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("fatal error: %s", err.Error())
	}
}

func configureDatabase(cfg config.Config, logger *slog.Logger) (*gorm.DB, error) {
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

	return db, nil
}

func dbDialector(cfg config.Config) gorm.Dialector {
	switch cfg.Database.Driver {
	case config.DatabaseDriverPostgres:
		return postgres.Open(cfg.Database.DSN)
	case config.DatabaseDriverSqlite:
		return sqlite.Open(cfg.Database.DSN)
	default:
		return nil
	}
}
