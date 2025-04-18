package objectstorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

func (s *LocalObjectStorage) SaveRemoteFile(ctx context.Context, remoteLocation, podcastGUID, fileName string) (int64, error) {
	err := util.ValidateExtURL(remoteLocation)
	if err != nil {
		return -1, fmt.Errorf("invalid remoteLocation '%s': %w", remoteLocation, err)
	}

	root, err := s.localRoot()
	if err != nil {
		return -1, err
	}

	err = mkdirIfNotExists(root, podcastGUID)
	if err != nil {
		return -1, err
	}

	localPath := path.Join(podcastGUID, fileName)
	f, err := root.Create(localPath)
	if err != nil {
		return -1, err
	}
	defer f.Close()

	req, err := http.NewRequest(http.MethodGet, remoteLocation, nil)
	if err != nil {
		return -1, err
	}
	req = req.WithContext(ctx)

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return -1, fmt.Errorf("failed to download file with status '%d'", resp.StatusCode)
	}

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return -1, err
	}

	return n, nil
}

func (s *LocalObjectStorage) ServeFile(ctx context.Context, r *http.Request, w http.ResponseWriter, podcastGUID, fileName string) error {
	root, err := s.localRoot()
	if err != nil {
		return err
	}

	filePath := path.Join(podcastGUID, fileName)
	f, err := root.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	http.ServeContent(w, r, "", time.Time{}, f)
	return nil
}

func (s *LocalObjectStorage) localRoot() (*os.Root, error) {
	err := os.MkdirAll(s.BasePath, 0750)
	if err != nil {
		return nil, err
	}
	return os.OpenRoot(s.BasePath)
}

func mkdirIfNotExists(root *os.Root, dir string) error {
	err := root.Mkdir(dir, 0750)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil
		}
		return err
	}
	return nil
}
