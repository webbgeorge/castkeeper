package objectstorage

import (
	"context"
	"io"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

type ObjectStorage interface {
	DownloadEpisodeFromSource(ctx context.Context, episode podcasts.Episode) error
	LoadEpisode(ctx context.Context, episode podcasts.Episode) (io.ReadSeekCloser, error)
	DownloadImageFromSource(ctx context.Context, podcast podcasts.Podcast) error
	LoadImage(ctx context.Context, podcast podcasts.Podcast) (io.ReadSeekCloser, error)
}
