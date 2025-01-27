package objectstorage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/util"
)

type LocalObjectStorage struct {
	HTTPClient *http.Client
	BasePath   string
}

func (s *LocalObjectStorage) SaveRemoteFile(ctx context.Context, remoteLocation, podcastGUID, fileName string) error {
	err := util.ValidateExtURL(remoteLocation)
	if err != nil {
		return fmt.Errorf("invalid remoteLocation '%s': %w", remoteLocation, err)
	}

	dir := path.Join(s.BasePath, podcastGUID)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	localPath := path.Join(dir, fileName)

	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	req, err := http.NewRequest(http.MethodGet, remoteLocation, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to download file with status '%d'", resp.StatusCode)
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *LocalObjectStorage) ServeFile(ctx context.Context, r *http.Request, w http.ResponseWriter, podcastGUID, fileName string) error {
	filePath := path.Join(s.BasePath, podcastGUID, fileName)
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	http.ServeContent(w, r, "", time.Time{}, f)
	return nil
}
