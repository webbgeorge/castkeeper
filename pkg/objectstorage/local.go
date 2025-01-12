package objectstorage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

type LocalObjectStorage struct {
	BasePath string
}

func (s *LocalObjectStorage) DownloadFromSource(episode podcasts.Episode) error {
	dir := path.Join(s.BasePath, episode.PodcastGUID)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s.%s", episode.GUID, podcasts.MimeToExt[episode.MimeType])

	f, err := os.Create(path.Join(dir, fileName))
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(episode.DownloadURL)
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
