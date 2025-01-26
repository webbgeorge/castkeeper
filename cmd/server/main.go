package main

import (
	"context"
	"fmt"
	"log"
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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, logger, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	// TODO remove when cfg is being used
	fmt.Println(cfg)

	gormLogger := slogGorm.New(
		slogGorm.WithHandler(logger.Handler()),
	)

	// TODO DB config from config file
	db, err := gorm.Open(
		sqlite.Open("test.db"),
		&gorm.Config{
			TranslateError: true,
			Logger:         gormLogger,
		},
	)
	if err != nil {
		log.Fatalf("failed to connect to database", err)
	}

	db.AutoMigrate(&podcasts.Podcast{})
	db.AutoMigrate(&podcasts.Episode{})

	objstore := &objectstorage.LocalObjectStorage{
		HTTPClient: framework.NewHTTPClient(time.Second * 5),
		BasePath:   "/Users/georgewebb/workspace/castkeeper/testout",
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
		dw := downloadworker.DownloadWorker{
			// _ = downloadworker.DownloadWorker{
			DB: db,
			OS: objstore,
		}
		return dw.Start(ctx)
		// return nil
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("fatal error: %s", err.Error())
	}
}
