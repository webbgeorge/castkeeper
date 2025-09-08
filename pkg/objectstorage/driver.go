package objectstorage

import (
	"context"
	"net/http"
)

type Credentials struct {
	Username string
	Password string
}

type ObjectStorage interface {
	SaveRemoteFile(ctx context.Context, creds *Credentials, remoteLocation, podcastGUID, fileName string) (int64, error)
	ServeFile(ctx context.Context, r *http.Request, w http.ResponseWriter, podcastGUID, fileName string) error
}
