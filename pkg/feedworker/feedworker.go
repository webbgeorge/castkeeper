package feedworker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

const (
	feedPollFrequency = time.Second * 10
	minCheckInterval  = time.Minute * 5
)

type FeedWorker struct {
	FeedService *podcasts.FeedService
	DB          *gorm.DB
}

func (w *FeedWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(feedPollFrequency)
	defer ticker.Stop()

	for {
		pods, err := podcasts.ListPodcasts(ctx, w.DB)
		if err != nil {
			framework.GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("feedworker failed to list podcasts: %s", err.Error()))
		}

		for _, pod := range pods {
			if pod.LastCheckedAt != nil && pod.LastCheckedAt.Add(minCheckInterval).After(time.Now()) {
				framework.GetLogger(ctx).DebugContext(ctx, fmt.Sprintf("podcast '%s' checked too recently, skipping", pod.GUID))
				continue
			}
			err := w.ProcessPodcast(ctx, pod)
			if err != nil {
				framework.GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("feedworker failed to process podcast '%s': %s", pod.GUID, err.Error()))
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
	_, episodes, err := w.FeedService.ParseFeed(ctx, podcast.FeedURL)
	if err != nil {
		if !errors.Is(err, podcasts.ParseErrors{}) {
			return err
		}
		framework.GetLogger(ctx).WarnContext(ctx, fmt.Sprintf("some episodes of podcast '%s' had parsing errors: %s", podcast.GUID, err.Error()))
		// continue even with some episode parse failures...
	}

	existingEpisodes, err := podcasts.ListEpisodes(ctx, w.DB, podcast.GUID)
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
