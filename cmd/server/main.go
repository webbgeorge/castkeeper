package main

import (
	"context"
	"log"
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
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
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

	otelRes, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(applicationName),
			semconv.DeploymentEnvironmentKey.String(cfg.EnvName),
			semconv.ServiceInstanceIDKey.String(framework.GetHostID()),
			semconv.ServiceVersionKey.String(castkeeper.Version),
		),
	)
	if err != nil {
		log.Fatalf("failed to create otel configuration", err)
	}

	otelLogger := otelslog.NewLogger("context")

	// TODO use otel slog logger
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
		return webserver.Start(ctx, otelRes, db, objstore, itunesAPI)
	})

	g.Go(func() error {
		fw := feedworker.FeedWorker{
			DB:     db,
			Logger: otelLogger,
		}
		return fw.Start(ctx)
	})

	g.Go(func() error {
		dw := downloadworker.DownloadWorker{
			DB:     db,
			OS:     objstore,
			Logger: otelLogger,
		}
		return dw.Start(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("fatal error: %s", err.Error())
	}
}
