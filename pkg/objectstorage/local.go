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

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/util"
)

type LocalObjectStorage struct {
	HTTPClient *http.Client
	Root       *os.Root
}

func (s *LocalObjectStorage) SaveRemoteFile(ctx context.Context, creds *podcasts.PodcastCredentials, remoteLocation, podcastGUID, fileName string) (int64, error) {
	err := util.ValidateExtURL(remoteLocation)
	if err != nil {
		return -1, fmt.Errorf("invalid remoteLocation '%s': %w", remoteLocation, err)
	}

	err = mkdirIfNotExists(s.Root, podcastGUID)
	if err != nil {
		return -1, err
	}

	localPath := path.Join(podcastGUID, fileName)
	f, err := s.Root.Create(localPath)
	if err != nil {
		return -1, err
	}
	defer f.Close()

	req, err := http.NewRequest(http.MethodGet, remoteLocation, nil)
	if err != nil {
		return -1, err
	}
	req = req.WithContext(ctx)

	if creds != nil {
		req.SetBasicAuth(creds.Username, creds.Password)
	}

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
	filePath := path.Join(podcastGUID, fileName)
	f, err := s.Root.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	http.ServeContent(w, r, "", time.Time{}, f)
	return nil
}

func MustOpenLocalFSRoot(basePath string) *os.Root {
	err := os.MkdirAll(basePath, 0750)
	if err != nil {
		panic(err)
	}
	root, err := os.OpenRoot(basePath)
	if err != nil {
		panic(err)
	}
	return root
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
