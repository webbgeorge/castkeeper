package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/webbgeorge/castkeeper"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/downloadworker"
	"github.com/webbgeorge/castkeeper/pkg/feedworker"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/webserver"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	applicationName = "castkeeper"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	slogger := framework.NewLogger(
		applicationName,
		cfg.EnvName,
		castkeeper.Version,
		slog.LevelInfo, // TODO level from config
	)

	// TODO use slog logger
	gormLog := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{LogLevel: logger.Warn, IgnoreRecordNotFoundError: true},
	)

	// TODO DB config from config file
	db, err := gorm.Open(
		sqlite.Open("test.db"),
		&gorm.Config{
			TranslateError: true,
			Logger:         gormLog,
		},
	)
	if err != nil {
		log.Fatalf("failed to connect to database", err)
	}

	db.AutoMigrate(&podcasts.Podcast{})
	db.AutoMigrate(&podcasts.Episode{})

	objstore := &objectstorage.LocalObjectStorage{
		BasePath: "/Users/georgewebb/workspace/castkeeper/testout",
	}

	itunesAPI := &itunes.ItunesAPI{
		HTTPClient: http.DefaultClient,
	}

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return webserver.Start(ctx, slogger, db, objstore, itunesAPI)
	})

	g.Go(func() error {
		fw := feedworker.FeedWorker{
			DB:     db,
			Logger: slogger,
		}
		return fw.Start(ctx)
	})

	g.Go(func() error {
		dw := downloadworker.DownloadWorker{
			// _ = downloadworker.DownloadWorker{
			DB:     db,
			OS:     objstore,
			Logger: slogger,
		}
		return dw.Start(ctx)
		// return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("fatal error: %s", err.Error())
	}
}
