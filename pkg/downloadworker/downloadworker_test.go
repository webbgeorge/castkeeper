package downloadworker_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/downloadworker"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

func TestDownloadWorker(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()
	root, resetFS := fixtures.ConfigureFSForTestWithFixtures()
	defer resetFS()

	dlWorker := downloadworker.NewDownloadWorkerQueueHandler(db, &objectstorage.LocalObjectStorage{
		HTTPClient: fixtures.TestDataHTTPClient,
		Root:       root,
	}, nil)

	// valid-eps-pending.xml fixture
	epGUID := fixtures.PodEpGUID("pending-ep-1")

	assertEpisodeStatus(db, t, epGUID, "pending")

	err := dlWorker(context.Background(), epGUID)

	assert.Nil(t, err)

	assertEpisodeStatus(db, t, epGUID, "success")
	assertEpisodeContent(db, root, t, epGUID, "ep1 content")
}

func TestDownloadWorker_InvalidQueueData(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()
	root, resetFS := fixtures.ConfigureFSForTestWithFixtures()
	defer resetFS()

	dlWorker := downloadworker.NewDownloadWorkerQueueHandler(db, &objectstorage.LocalObjectStorage{
		HTTPClient: fixtures.TestDataHTTPClient,
		Root:       root,
	}, nil)

	err := dlWorker(context.Background(), nil)

	assert.Equal(t, "failed to get episodeGUID from queue data", err.Error())
}

func TestDownloadWorker_EpisodeNotFound(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()
	root, resetFS := fixtures.ConfigureFSForTestWithFixtures()
	defer resetFS()

	dlWorker := downloadworker.NewDownloadWorkerQueueHandler(db, &objectstorage.LocalObjectStorage{
		HTTPClient: fixtures.TestDataHTTPClient,
		Root:       root,
	}, nil)

	err := dlWorker(context.Background(), "not-an-ep")

	assert.Equal(t, "failed to get a pending episode: record not found", err.Error())
}

func TestDownloadWorker_FailedToDownload(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()
	root, resetFS := fixtures.ConfigureFSForTestWithFixtures()
	defer resetFS()

	dlWorker := downloadworker.NewDownloadWorkerQueueHandler(db, &objectstorage.LocalObjectStorage{
		HTTPClient: fixtures.TestDataHTTPClient,
		Root:       root,
	}, nil)

	if err := db.Create(&podcasts.Episode{
		GUID:        "test-download-failure",
		PodcastGUID: "pod-guid-eg",
		Title:       "Test",
		DownloadURL: "http://testdata/error",
		MimeType:    "audio/mpeg",
		Status:      "pending",
	}).Error; err != nil {
		panic(err)
	}

	err := dlWorker(context.Background(), "test-download-failure")

	assert.Equal(t, "failed to download episode 'test-download-failure': failed to download file with status '500'", err.Error())
}

func assertEpisodeStatus(db *gorm.DB, t *testing.T, episodeGUID, expectedStatus string) {
	t.Helper()
	ep, err := podcasts.GetEpisode(context.Background(), db, episodeGUID)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectedStatus, ep.Status)
}

func assertEpisodeContent(db *gorm.DB, root *os.Root, t *testing.T, episodeGUID, expectedContent string) {
	t.Helper()
	ep, err := podcasts.GetEpisode(context.Background(), db, episodeGUID)
	if err != nil {
		panic(err)
	}
	f, err := root.Open(fmt.Sprintf("%s/%s.mp3", ep.PodcastGUID, ep.GUID))
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectedContent, strings.TrimSpace(string(data)))
}
