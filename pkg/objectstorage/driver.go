package objectstorage

import (
	"io"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

type ObjectStorage interface {
	DownloadEpisodeFromSource(episode podcasts.Episode) error
	LoadEpisode(episode podcasts.Episode) (io.ReadSeekCloser, error)
	DownloadImageFromSource(podcast podcasts.Podcast) error
	LoadImage(podcast podcasts.Podcast) (io.ReadSeekCloser, error)
}
