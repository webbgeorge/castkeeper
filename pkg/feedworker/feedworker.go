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
	FeedWorkerQueueName = "feedWorker"
	minCheckInterval    = time.Minute * 10
)

func NewFeedWorkerQueueHandler(db *gorm.DB, feedService *podcasts.FeedService) func(context.Context, any) error {
	return func(ctx context.Context, _ any) error {
		pods, err := podcasts.ListPodcasts(ctx, db)
		if err != nil {
			framework.GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("feedworker failed to list podcasts: %s", err.Error()))
			return err
		}

		errs := make([]error, 0)
		for _, pod := range pods {
			// TODO better logic for if should check
			if pod.LastCheckedAt != nil && pod.LastCheckedAt.Add(minCheckInterval).After(time.Now()) {
				framework.GetLogger(ctx).DebugContext(ctx, fmt.Sprintf("podcast '%s' checked too recently, skipping", pod.GUID))
				continue
			}
			err := processPodcast(ctx, db, feedService, pod)
			if err != nil {
				framework.GetLogger(ctx).ErrorContext(ctx, fmt.Sprintf("feedworker failed to process podcast '%s': %s", pod.GUID, err.Error()))
				errs = append(errs, err)
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		return nil
	}
}

func processPodcast(ctx context.Context, db *gorm.DB, feedService *podcasts.FeedService, podcast podcasts.Podcast) error {
	_, episodes, err := feedService.ParseFeed(ctx, podcast.FeedURL)
	if err != nil {
		if !errors.Is(err, podcasts.ParseErrors{}) {
			return err
		}
		framework.GetLogger(ctx).WarnContext(ctx, fmt.Sprintf("some episodes of podcast '%s' had parsing errors: %s", podcast.GUID, err.Error()))
		// continue even with some episode parse failures...
	}

	existingEpisodes, err := podcasts.ListEpisodes(ctx, db, podcast.GUID)
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

		if err := db.Create(&ep).Error; err != nil {
			return err
		}
	}

	now := time.Now()
	var lastEpisodeAt *time.Time
	if len(episodes) > 0 {
		// feed items are sorted oldest to newest
		lastEpisodeAt = &episodes[len(episodes)-1].PublishedAt
	}

	err = podcasts.UpdatePodcastTimes(ctx, db, &podcast, &now, lastEpisodeAt)
	if err != nil {
		return err
	}

	return nil
}
