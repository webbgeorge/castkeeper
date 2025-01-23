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

func (s *LocalObjectStorage) DownloadEpisodeFromSource(episode podcasts.Episode) error {
	dir := path.Join(s.BasePath, episode.PodcastGUID)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s.%s", episode.GUID, podcasts.MimeToExt[episode.MimeType])

	return s.downloadFile(episode.DownloadURL, path.Join(dir, fileName))
}

func (s *LocalObjectStorage) LoadEpisode(episode podcasts.Episode) (io.ReadSeekCloser, error) {
	filePath := path.Join(s.BasePath, episode.PodcastGUID, fmt.Sprintf("%s.%s", episode.GUID, podcasts.MimeToExt[episode.MimeType]))
	return os.Open(filePath)
}

func (s *LocalObjectStorage) DownloadImageFromSource(podcast podcasts.Podcast) error {
	dir := path.Join(s.BasePath, podcast.GUID)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	// TODO detect extension
	fileName := fmt.Sprintf("%s.%s", podcast.GUID, ".jpg")

	return s.downloadFile(podcast.ImageURL, path.Join(dir, fileName))
}

func (s *LocalObjectStorage) LoadImage(podcast podcasts.Podcast) (io.ReadSeekCloser, error) {
	// TODO detect extension
	filePath := path.Join(s.BasePath, podcast.GUID, fmt.Sprintf("%s.%s", podcast.GUID, ".jpg"))
	return os.Open(filePath)
}

func (s *LocalObjectStorage) downloadFile(remoteLocation, localPath string) error {
	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(remoteLocation)
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
