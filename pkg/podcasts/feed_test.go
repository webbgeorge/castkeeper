package podcasts_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
)

func TestParseFeed(t *testing.T) {
	testCases := map[string]struct {
		url              string
		expectedPodcast  podcasts.Podcast
		expectedEpisodes []podcasts.Episode
		expectedErr      string
	}{
		"invalid external URL": {
			url:              "http://localhost/feed.xml",
			expectedPodcast:  podcasts.Podcast{},
			expectedEpisodes: nil,
			expectedErr:      "invalid feedURL 'http://localhost/feed.xml': URL host must not be localhost",
		},
		"invalid feed": {
			url:              "http://testdata/feeds/invalid.xml",
			expectedPodcast:  podcasts.Podcast{},
			expectedEpisodes: nil,
			expectedErr:      "failed to parse feed: EOF",
		},
		"valid feed": {
			url: "http://testdata/feeds/valid.xml",
			expectedPodcast: fakePodcast(
				"http://testdata/feeds/valid.xml",
				fixtures.PodEpGUID("abc-123"),
				timePtrStr("2024-12-27T11:12:13"),
			),
			expectedEpisodes: []podcasts.Episode{
				fakeEpisode(fixtures.PodEpGUID("ep-1"), fixtures.PodEpGUID("abc-123"), timeFromStr("2024-12-26T11:12:13")),
				fakeEpisode(fixtures.PodEpGUID("ep-2"), fixtures.PodEpGUID("abc-123"), timeFromStr("2024-12-27T11:12:13")),
			},
			expectedErr: "",
		},
		"no eps": {
			url: "http://testdata/feeds/no-eps.xml",
			expectedPodcast: fakePodcast(
				"http://testdata/feeds/no-eps.xml",
				fixtures.PodEpGUID("abc-123"),
				nil,
			),
			expectedEpisodes: []podcasts.Episode{},
			expectedErr:      "",
		},
		"podcast guid fallback": {
			url: "http://testdata/feeds/no-pod-guid.xml",
			expectedPodcast: fakePodcast(
				"http://testdata/feeds/no-pod-guid.xml",
				fixtures.PodEpGUID("http://www.example.com/feed"), // generated fallback GUID
				nil,
			),
			expectedEpisodes: []podcasts.Episode{},
			expectedErr:      "",
		},
		"episode with no url gives error": {
			url: "http://testdata/feeds/ep-no-url.xml",
			expectedPodcast: fakePodcast(
				"http://testdata/feeds/ep-no-url.xml",
				"abc-123",
				nil,
			),
			expectedEpisodes: []podcasts.Episode{},
			expectedErr:      fmt.Sprintf("1 errors whilst parsing episodes: could not read download URL, skipping episode '%s'", fixtures.PodEpGUID("ep-1")),
		},
		"episode with invalid file type gives error": {
			url: "http://testdata/feeds/invalid-mime.xml",
			expectedPodcast: fakePodcast(
				"http://testdata/feeds/invalid-mime.xml",
				"abc-123",
				nil,
			),
			expectedEpisodes: []podcasts.Episode{},
			expectedErr:      fmt.Sprintf("1 errors whilst parsing episodes: unsupported file type 'not/type', skipping episode '%s'", fixtures.PodEpGUID("ep-1")),
		},
	}

	// intercepts HTTP requests and returns test data based on the URL
	feedService := podcasts.FeedService{
		HTTPClient: fixtures.TestDataHTTPClient,
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			podcast, episodes, err := feedService.ParseFeed(context.Background(), tc.url)

			if err == nil {
				assert.Equal(t, tc.expectedPodcast, podcast)
				assert.Equal(t, tc.expectedEpisodes, episodes)
			} else {
				assert.Equal(t, tc.expectedErr, err.Error())
			}
		})
	}
}

func TestParseFeedTruncation(t *testing.T) {
	// intercepts HTTP requests and returns test data based on the URL
	feedService := podcasts.FeedService{
		HTTPClient: fixtures.TestDataHTTPClient,
	}

	podcast, episodes, err := feedService.ParseFeed(
		context.Background(),
		"http://testdata/feeds/very-long-pod-title.xml",
	)

	assert.Nil(t, err)
	assert.Len(t, podcast.Title, 500)
	assert.Len(t, episodes, 1)
	assert.Len(t, episodes[0].Title, 500)
}

func TestParseFeedEpisodeGUIDFallback(t *testing.T) {
	// intercepts HTTP requests and returns test data based on the URL
	feedService := podcasts.FeedService{
		HTTPClient: fixtures.TestDataHTTPClient,
	}

	_, episodes, err := feedService.ParseFeed(
		context.Background(),
		"http://testdata/feeds/no-ep-guid.xml",
	)

	assert.Nil(t, err)
	assert.Len(t, episodes, 1)
	assert.Equal(t, fixtures.PodEpGUID("Test episode with no guid2024-12-26T11:12:13Z"), episodes[0].GUID)
}

func TestGenerateFeed(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	feed, err := podcasts.GenerateFeed(context.Background(), "http://example.com", db, fixtures.PodEpGUID("abc-123"))
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	feed.WriteFeedXML(buf)

	exp, err := os.ReadFile("testdata/expected-generated-feed.xml")
	if err != nil {
		panic(err)
	}

	assert.Equal(
		t,
		strings.TrimSpace(string(exp)),
		strings.TrimSpace(buf.String()),
	)
}

func timeFromStr(tStr string) time.Time {
	t, _ := time.Parse("2006-01-02T15:04:05", tStr)
	return t
}

func timePtrStr(tStr string) *time.Time {
	t := timeFromStr(tStr)
	return &t
}

func fakePodcast(feedURL, guid string, latestEpPubAt *time.Time) podcasts.Podcast {
	return podcasts.Podcast{
		GUID:        guid,
		Title:       fmt.Sprintf("Test podcast %s", guid),
		Author:      "Dr Tester",
		Description: "Test podcast description goes here",
		Language:    "en",
		Link:        "http://www.example.com/podcast-site",
		Categories: []podcasts.Category{
			{Name: "Comedy"},
			{Name: "Drama", SubCategory: &podcasts.Category{Name: "Thriller"}},
		},
		IsExplicit:    true,
		ImageURL:      "http://www.example.com/image.jpg",
		FeedURL:       feedURL,
		LastCheckedAt: nil,
		LastEpisodeAt: latestEpPubAt,
	}
}

func fakeEpisode(guid, podGuid string, pubAt time.Time) podcasts.Episode {
	return podcasts.Episode{
		GUID:         guid,
		PodcastGUID:  podGuid,
		Title:        fmt.Sprintf("Test episode %s", guid),
		Description:  "Episode test description",
		DownloadURL:  fmt.Sprintf("http://www.example.com/episode-%s.mp3", guid),
		MimeType:     "audio/mpeg",
		DurationSecs: 1234,
		PublishedAt:  pubAt,
	}
}
