package objectstorage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

type LocalObjectStorage struct {
	BasePath string
}

func (s *LocalObjectStorage) SaveFromURL(url, podcastGUID, episodeGUID string) error {
	dir := path.Join(s.BasePath, podcastGUID)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(dir, episodeGUID))
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to download file with status '%d'", resp.StatusCode)
	}

	// TODO verify it is an audio file?

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
