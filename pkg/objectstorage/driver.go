package objectstorage

import "github.com/webbgeorge/castkeeper/pkg/podcasts"

type ObjectStorage interface {
	DownloadFromSource(episode podcasts.Episode) error
	// Load(episode podcasts.Episode) (io.Reader, error)
}
