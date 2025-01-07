package downloadworker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

type DownloadWorker struct {
	DB *gorm.DB
	OS objectstorage.ObjectStorage
}

func (w *DownloadWorker) Start(ctx context.Context) error {
	for {
		// handle cancellation at top to ensure it runs on every iteration
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := w.ProcessEpisode(ctx)
		if err != nil {
			if errors.As(err, &podcasts.ErrEpisodeNotFound) {
				time.Sleep(5 * time.Second)
				// don't log if no eps in queue
				fmt.Println("not found", err) // TODO
				continue
			}

			// TODO log err
			fmt.Println("err processing episode", err)

			// small sleep to avoid hammering DB on repeated errs
			time.Sleep(time.Second)
			continue
		}

		// TODO log success
		fmt.Println("success downloading episode")
	}
}

func (w *DownloadWorker) ProcessEpisode(ctx context.Context) error {
	episode, err := podcasts.GetPendingEpisode(ctx, w.DB)
	if err != nil {
		return fmt.Errorf("failed to get a pending episode: %w", err)
	}

	podcast, err := podcasts.GetPodcast(ctx, w.DB, episode.PodcastID)
	if err != nil {
		return fmt.Errorf("failed to fetch podcast for episode '%d': %w", episode.ID, err)
	}

	err = w.ProcessDownload(ctx, episode, podcast)
	if err != nil {
		// TODO allow for multiple failed attempts instead of immediately failing
		upErr := podcasts.UpdateEpisodeStatus(ctx, w.DB, &episode, podcasts.EpisodeStatusFailed)
		if upErr != nil {
			return fmt.Errorf("failed to update episode '%d' status to failed: %w", episode.ID, upErr)
		}
		return fmt.Errorf("failed to download episode '%d': %w", episode.ID, err)
	}

	err = podcasts.UpdateEpisodeStatus(ctx, w.DB, &episode, podcasts.EpisodeStatusSuccess)
	if err != nil {
		return fmt.Errorf("failed to update episode '%s' status to success: %w", episode.ID, err)
	}

	return nil
}

func (w *DownloadWorker) ProcessDownload(ctx context.Context, episode podcasts.Episode, podcast podcasts.Podcast) error {
	return w.OS.SaveFromURL(episode.DownloadURL, podcast.GUID, episode.GUID)
}
