package downloadworker

import (
	"context"
	"errors"
	"fmt"

	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/util"
	"gorm.io/gorm"
)

const DownloadWorkerQueueName = "downloadWorker"

func NewDownloadWorkerQueueHandler(db *gorm.DB, os objectstorage.ObjectStorage) func(context.Context, any) error {
	return func(ctx context.Context, episodeGUIDAny any) error {
		episodeGUID, ok := episodeGUIDAny.(string)
		if !ok {
			return errors.New("failed to get episodeGUID from queue data")
		}

		episode, err := podcasts.GetEpisode(ctx, db, episodeGUID)
		if err != nil {
			return fmt.Errorf("failed to get a pending episode: %w", err)
		}

		fileName := fmt.Sprintf("%s.%s", util.SanitiseGUID(episode.GUID), podcasts.MimeToExt[episode.MimeType])
		n, err := os.SaveRemoteFile(ctx, episode.DownloadURL, util.SanitiseGUID(episode.PodcastGUID), fileName)
		if err != nil {
			upErr := podcasts.UpdateEpisodeStatus(ctx, db, &episode, podcasts.EpisodeStatusFailed, nil)
			if upErr != nil {
				return fmt.Errorf("failed to update episode '%s' status to failed: %w", episode.GUID, upErr)
			}
			return fmt.Errorf("failed to download episode '%s': %w", episode.GUID, err)
		}

		err = podcasts.UpdateEpisodeStatus(ctx, db, &episode, podcasts.EpisodeStatusSuccess, &n)
		if err != nil {
			return fmt.Errorf("failed to update episode '%s' status to success: %w", episode.GUID, err)
		}

		return nil
	}
}
