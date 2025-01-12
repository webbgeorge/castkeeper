package feedworker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

// TODO config? reduce this to 10 secs once lastchecked logic is in
const feedPollFrequency = time.Minute * 5

type FeedWorker struct {
	DB     *gorm.DB
	Logger *slog.Logger
}

func (w *FeedWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(feedPollFrequency)
	defer ticker.Stop()

	for {
		pods, err := podcasts.ListPodcasts(ctx, w.DB)
		if err != nil {
			w.Logger.ErrorContext(ctx, fmt.Sprintf("feedworker failed to list podcasts: %s", err.Error()))
		}

		for _, pod := range pods {
			// TODO check last checked time
			err := w.ProcessPodcast(ctx, pod)
			if err != nil {
				w.Logger.ErrorContext(ctx, fmt.Sprintf("feedworker failed to process podcast '%s': %s", pod.GUID, err.Error()))
			}
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (w *FeedWorker) ProcessPodcast(ctx context.Context, podcast podcasts.Podcast) error {
	feed, err := podcasts.ParseFeed(ctx, podcast.FeedURL)
	if err != nil {
		return err
	}
	episodes, errs := podcasts.EpisodesFromFeed(feed)

	if len(errs) > 0 {
		w.Logger.WarnContext(ctx, fmt.Sprintf("some episodes of podcast '%s' had parsing errors: %v", podcast.GUID, errs))
	}

	existingEpisodes, err := podcasts.ListEpisodes(ctx, w.DB)
	if err != nil {
		return err
	}

	for _, ep := range episodes {
		exists := false
		for _, exEp := range existingEpisodes {
			if exEp.GUID == ep.GUID {
				exists = true
				break
			}
		}

		if exists {
			continue
		}

		ep.PodcastGUID = podcast.GUID
		ep.Status = podcasts.EpisodeStatusPending

		if err := w.DB.Create(&ep).Error; err != nil {
			return err
		}
	}

	now := time.Now()
	var lastEpisodeAt *time.Time
	if len(episodes) > 0 {
		// feed items are sorted oldest to newest
		lastEpisodeAt = &episodes[len(episodes)-1].PublishedAt
	}

	err = podcasts.UpdatePodcastTimes(ctx, w.DB, &podcast, &now, lastEpisodeAt)
	if err != nil {
		return err
	}

	return nil
}
