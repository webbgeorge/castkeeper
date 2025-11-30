package objectstorage

import (
	"context"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

type ObjectStorage interface {
	SaveRemoteFile(ctx context.Context, creds *podcasts.PodcastCredentials, remoteLocation, podcastGUID, fileName string) (int64, error)
	ServeFile(ctx context.Context, r *http.Request, w http.ResponseWriter, podcastGUID, fileName string) error
}
