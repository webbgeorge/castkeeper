package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/downloadworker"
	"github.com/webbgeorge/castkeeper/pkg/feedworker"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/webserver"
	"golang.org/x/sync/errgroup"
)

func main() {
	configFile := "" // optional specific config file (otherwise uses default locations)
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	cfg, logger, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	ctx := framework.ContextWithLogger(context.Background(), logger)

	db, err := database.ConfigureDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	objstore, err := configureObjectStorage(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to configure objectstorage: %v", err)
	}

	feedService := &podcasts.FeedService{
		HTTPClient: framework.NewHTTPClient(time.Second * 5),
	}
	itunesAPI := &itunes.ItunesAPI{
		HTTPClient: framework.NewHTTPClient(time.Second * 5),
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return webserver.Start(ctx, cfg, logger, feedService, db, objstore, itunesAPI)
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
			DB: db,
			OS: objstore,
		}
		return dw.Start(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("fatal error: %s", err.Error())
	}
}

func configureObjectStorage(ctx context.Context, cfg config.Config) (objectstorage.ObjectStorage, error) {
	httpClient := framework.NewHTTPClient(time.Minute * 15)

	switch cfg.ObjectStorage.Driver {
	case config.ObjectStorageDriverLocal:
		return &objectstorage.LocalObjectStorage{
			HTTPClient: httpClient,
			BasePath:   cfg.ObjectStorage.LocalBasePath,
		}, nil

	case config.ObjectStorageDriverS3:
		// uses aws environment variables to configure the SDK
		awsCfg, err := awsConfig.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
		s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.UsePathStyle = cfg.ObjectStorage.S3ForcePathStyle
		})

		return &objectstorage.S3ObjectStorage{
			HTTPClient: httpClient,
			S3Client:   s3Client,
			BucketName: cfg.ObjectStorage.S3Bucket,
			Prefix:     cfg.ObjectStorage.S3Prefix,
		}, nil

	default:
		return nil, errors.New("unknown objectstorage driver")
	}
}
