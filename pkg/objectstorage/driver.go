package objectstorage

import (
	"context"
	"net/http"
)

type ObjectStorage interface {
	SaveRemoteFile(ctx context.Context, remoteLocation, podcastGUID, fileName string) error
	ServeFile(ctx context.Context, r *http.Request, w http.ResponseWriter, podcastGUID, fileName string) error
}
