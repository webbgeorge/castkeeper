package downloadworker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

type DownloadWorker struct {
	DB     *gorm.DB
	OS     objectstorage.ObjectStorage
	Logger *slog.Logger
}

func (w *DownloadWorker) Start(ctx context.Context) error {
	for {
		// handle cancellation at top to ensure it runs on every iteration
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		episode, err := w.ProcessEpisode(ctx)
		if err != nil {
			notFoundErr := podcasts.ErrEpisodeNotFound // copy first to avoid changing value
			if errors.As(err, &notFoundErr) {
				time.Sleep(5 * time.Second) // no jobs on queue, wait before next poll
				continue
			}

			w.Logger.ErrorContext(ctx, fmt.Sprintf("downloadworker failed to process episode: %s", err.Error()))

			// small sleep to avoid hammering DB on repeated errs
			time.Sleep(time.Second)
			continue
		}

		w.Logger.InfoContext(ctx, fmt.Sprintf("successfully downloaded episode '%s'", episode.GUID))
	}
}

func (w *DownloadWorker) ProcessEpisode(ctx context.Context) (*podcasts.Episode, error) {
	episode, err := podcasts.GetPendingEpisode(ctx, w.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to get a pending episode: %w", err)
	}

	err = w.OS.DownloadFromSource(episode)
	if err != nil {
		// TODO allow for multiple failed attempts instead of immediately failing
		upErr := podcasts.UpdateEpisodeStatus(ctx, w.DB, &episode, podcasts.EpisodeStatusFailed)
		if upErr != nil {
			return nil, fmt.Errorf("failed to update episode '%d' status to failed: %w", episode.GUID, upErr)
		}
		return nil, fmt.Errorf("failed to download episode '%d': %w", episode.GUID, err)
	}

	err = podcasts.UpdateEpisodeStatus(ctx, w.DB, &episode, podcasts.EpisodeStatusSuccess)
	if err != nil {
		return nil, fmt.Errorf("failed to update episode '%s' status to success: %w", episode.GUID, err)
	}

	return &episode, nil
}
