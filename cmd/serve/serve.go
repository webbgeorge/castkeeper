package serve

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/webbgeorge/castkeeper/pkg/auth/sessions"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
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
	Use:     "serve",
	Short:   "Start the CastKeeper server",
	GroupID: "commands",
	Run:     run,
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

	objstore, err := objectstorage.ConfigureObjectStorage(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to configure objectstorage: %v", err)
	}

	encService, err := encryption.ConfigureEncryptedValueService(cfg)
	if err != nil {
		log.Fatalf("failed to configure encryption: %v", err)
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
			NewWebserver(cfg, logger, feedService, db, objstore, itunesAPI, encService).
			Start(ctx)
	})

	g.Go(func() error {
		scheduler := framework.TaskScheduler{
			DB: db,
			Tasks: []framework.ScheduledTaskDefinition{
				{TaskName: feedworker.FeedWorkerQueueName, Interval: time.Minute},
				{TaskName: sessions.HouseKeepingQueueName, Interval: time.Hour},
			},
		}
		return scheduler.Start(ctx)
	})

	g.Go(func() error {
		qw := framework.QueueWorker{
			DB:        db,
			QueueName: feedworker.FeedWorkerQueueName,
			HandlerFn: feedworker.NewFeedWorkerQueueHandler(db, feedService, encService),
		}
		return qw.Start(ctx)
	})

	g.Go(func() error {
		qw := framework.QueueWorker{
			DB:        db,
			QueueName: downloadworker.DownloadWorkerQueueName,
			HandlerFn: downloadworker.NewDownloadWorkerQueueHandler(db, objstore, encService),
		}
		return qw.Start(ctx)
	})

	g.Go(func() error {
		qw := framework.QueueWorker{
			DB:        db,
			QueueName: sessions.HouseKeepingQueueName,
			HandlerFn: sessions.NewHouseKeepingQueueWorker(db),
		}
		return qw.Start(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Fatalf("fatal error: %s", err.Error())
	}
}
