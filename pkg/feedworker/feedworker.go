package feedworker

import (
	"context"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

// TODO config? also, this is probably too frequent
const feedPollFrequency = time.Minute * 5

type FeedWorker struct {
	DB *gorm.DB
}

func (w *FeedWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(feedPollFrequency)
	defer ticker.Stop()

	for {
		pods, err := podcasts.ListPodcasts(ctx, w.DB)
		if err != nil {
			// TODO log failure and continue loop
		}

		for _, pod := range pods {
			err := w.ProcessPodcast(ctx, pod)
			if err != nil {
				// TODO log failure and continue loop
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
	episodes := podcasts.EpisodesFromFeed(feed)

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

		ep.PodcastID = podcast.ID
		ep.Status = podcasts.EpisodeStatusPending

		if err := w.DB.Create(&ep).Error; err != nil {
			return err
		}
	}

	return nil
}
