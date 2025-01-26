package objectstorage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

type LocalObjectStorage struct {
	HTTPClient *http.Client
	BasePath   string
}

func (s *LocalObjectStorage) DownloadEpisodeFromSource(ctx context.Context, episode podcasts.Episode) error {
	dir := path.Join(s.BasePath, episode.PodcastGUID)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s.%s", episode.GUID, podcasts.MimeToExt[episode.MimeType])

	return s.downloadFile(ctx, episode.DownloadURL, path.Join(dir, fileName))
}

func (s *LocalObjectStorage) LoadEpisode(ctx context.Context, episode podcasts.Episode) (io.ReadSeekCloser, error) {
	filePath := path.Join(s.BasePath, episode.PodcastGUID, fmt.Sprintf("%s.%s", episode.GUID, podcasts.MimeToExt[episode.MimeType]))
	return os.Open(filePath)
}

func (s *LocalObjectStorage) DownloadImageFromSource(ctx context.Context, podcast podcasts.Podcast) error {
	dir := path.Join(s.BasePath, podcast.GUID)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	// TODO detect extension
	fileName := fmt.Sprintf("%s.%s", podcast.GUID, ".jpg")

	return s.downloadFile(ctx, podcast.ImageURL, path.Join(dir, fileName))
}

func (s *LocalObjectStorage) LoadImage(ctx context.Context, podcast podcasts.Podcast) (io.ReadSeekCloser, error) {
	// TODO detect extension
	filePath := path.Join(s.BasePath, podcast.GUID, fmt.Sprintf("%s.%s", podcast.GUID, ".jpg"))
	return os.Open(filePath)
}

func (s *LocalObjectStorage) downloadFile(ctx context.Context, remoteLocation, localPath string) error {
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
