package podcasts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

func TestAddPodcast_Success(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	feedURL := "http://testdata/feeds/valid-not-added.xml"
	pod, err := podcasts.AddPodcast(
		context.Background(), db, feedService(), evs(), feedURL, nil)

	assert.Nil(t, err)
	assert.Equal(t, "Test podcast 2", pod.Title)

	// assert pod was saved to DB
	dbPod, err := podcasts.GetPodcast(context.Background(), db, pod.GUID)
	assert.Nil(t, err)
	assert.Equal(t, "Test podcast 2", dbPod.Title)
}

func TestAddPodcast_SuccessAuthenticatedFeed(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	feedURL := "http://testdata/authenticated/feeds/valid-not-added.xml"
	pod, err := podcasts.AddPodcast(
		context.Background(), db, feedService(),
		evs(), feedURL, &fixtures.AuthenticatedFeedCreds)

	assert.Nil(t, err)
	assert.Equal(t, "Test authenticated podcast 2", pod.Title)

	// assert pod was saved to DB
	dbPod, err := podcasts.GetPodcast(context.Background(), db, pod.GUID)
	assert.Nil(t, err)
	assert.Equal(t, "Test authenticated podcast 2", dbPod.Title)
}

func TestAddPodcast_InvalidFeed(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	feedURL := "http://testdata/feeds/invalid.xml"
	_, err := podcasts.AddPodcast(
		context.Background(), db, feedService(), evs(), feedURL, nil)

	assert.Equal(t, "failed to parse feed: EOF", err.Error())
}

func TestAddPodcast_InvalidCredentials(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	feedURL := "http://testdata/authenticated/feeds/valid-not-added.xml"
	invalidCreds := podcasts.PodcastCredentials{
		Username: "invalid",
		Password: "invalid",
	}
	_, err := podcasts.AddPodcast(
		context.Background(), db, feedService(),
		evs(), feedURL, &invalidCreds)

	assert.Equal(t, "failed to parse feed: non-200 http response '401'", err.Error())
}

func TestGetCredentials(t *testing.T) {
	feedURL := "http://example.com/feed"
	encCreds, err := evs().Encrypt(
		[]byte(`{"Username":"testuser","Password":"testpass"}`), []byte(feedURL))
	if err != nil {
		panic(err)
	}
	pod := podcasts.Podcast{
		FeedURL:     feedURL,
		Credentials: &encCreds,
	}

	creds, err := podcasts.GetCredentials(evs(), pod)
	assert.Nil(t, err)
	assert.NotNil(t, creds)
	assert.Equal(t, "testuser", creds.Username)
	assert.Equal(t, "testpass", creds.Password)
}

func TestGetCredentials_NoCreds(t *testing.T) {
	pod := podcasts.Podcast{
		FeedURL:     "http://example.com/feed",
		Credentials: nil,
	}
	creds, err := podcasts.GetCredentials(evs(), pod)
	assert.Nil(t, err)
	assert.Nil(t, creds)
}

func TestGetCredentials_FailedToDecrypt(t *testing.T) {
	encCreds, err := evs().Encrypt(
		[]byte(`{"Username":"testuser","Password":"testpass"}`),
		[]byte("http://example.com/feed"),
	)
	if err != nil {
		panic(err)
	}
	pod := podcasts.Podcast{
		FeedURL:     "http://example.com/modified-feed-url", // will cause decrypt to fail
		Credentials: &encCreds,
	}

	_, err = podcasts.GetCredentials(evs(), pod)
	assert.Equal(t, "aes_gcm_siv: message authentication failure", err.Error())
}

func evs() *encryption.EncryptedValueService {
	return fixtures.ConfigureEncryptedValueServiceForTest()
}

func feedService() *podcasts.FeedService {
	return &podcasts.FeedService{
		HTTPClient: fixtures.TestDataHTTPClient,
	}
}
