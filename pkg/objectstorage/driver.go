package objectstorage

import (
	"io"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

type ObjectStorage interface {
	DownloadFromSource(episode podcasts.Episode) error
	Load(episode podcasts.Episode) (io.ReadSeekCloser, error)
}
