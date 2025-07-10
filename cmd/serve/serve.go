package serve

import (
	"context"
	"errors"
	"log"
	"path"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth"
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

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the CastKeeper server",
	Long:  "TODO",
	Run:   run,
}

var cfgFile string

func init() {
	ServeCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (otherwise uses default locations)")
}

func run(cmd *cobra.Command, args []string) {
	cfg, logger, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	ctx := framework.ContextWithLogger(context.Background(), logger)

	db, err := database.ConfigureDatabase(cfg, logger, false)
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
		return webserver.
			NewWebserver(cfg, logger, feedService, db, objstore, itunesAPI).
			Start(ctx)
	})

	g.Go(func() error {
		scheduler := framework.TaskScheduler{
			DB: db,
			Tasks: []framework.ScheduledTaskDefinition{
				{TaskName: feedworker.FeedWorkerQueueName, Interval: time.Minute},
				{TaskName: auth.HouseKeepingQueueName, Interval: time.Hour},
			},
		}
		return scheduler.Start(ctx)
	})

	g.Go(func() error {
		qw := framework.QueueWorker{
			DB:        db,
			QueueName: feedworker.FeedWorkerQueueName,
			HandlerFn: feedworker.NewFeedWorkerQueueHandler(db, feedService),
		}
		return qw.Start(ctx)
	})

	g.Go(func() error {
		qw := framework.QueueWorker{
			DB:        db,
			QueueName: downloadworker.DownloadWorkerQueueName,
			HandlerFn: downloadworker.NewDownloadWorkerQueueHandler(db, objstore),
		}
		return qw.Start(ctx)
	})

	g.Go(func() error {
		qw := framework.QueueWorker{
			DB:        db,
			QueueName: auth.HouseKeepingQueueName,
			HandlerFn: auth.NewHouseKeepingQueueWorker(db),
		}
		return qw.Start(ctx)
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
			Root: objectstorage.MustOpenLocalFSRoot(
				path.Join(cfg.DataPath, "objects"),
			),
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
