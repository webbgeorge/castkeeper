package podcasts

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedService struct {
	HTTPClient *http.Client
}

func (s *FeedService) ParseFeed(ctx context.Context, feedURL string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	fp.Client = s.HTTPClient

	feed, err := fp.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	if feed.ITunesExt == nil {
		return nil, errors.New("feed is not a podcast")
	}

	// oldest to newest
	sort.Sort(feed)

	return feed, nil
}

func PodcastFromFeed(feedURL string, feed *gofeed.Feed) Podcast {
	var lastEpisodeAt *time.Time
	episodes, _ := EpisodesFromFeed(feed)
	if len(episodes) > 0 {
		// feed items are sorted oldest to newest
		lastEpisodeAt = &episodes[len(episodes)-1].PublishedAt
	}

	author := "unknown"
	if feed.Authors != nil && len(feed.Authors) > 0 && feed.Authors[0] != nil {
		author = feed.Authors[0].Name
	}

	imageURL := ""
	if feed.Image != nil && feed.Image.URL != "" {
		imageURL = feed.Image.URL
	}

	podcast := Podcast{
		GUID:          feedGUID(feed),
		Title:         truncate(feed.Title, 1000),
		Author:        author,
		Description:   truncate(feed.Description, 10000),
		ImageURL:      imageURL,
		FeedURL:       feedURL,
		LastCheckedAt: nil,
		LastEpisodeAt: lastEpisodeAt,
	}

	return podcast
}

func EpisodesFromFeed(feed *gofeed.Feed) ([]Episode, []error) {
	episodes := make([]Episode, 0)
	errs := make([]error, 0)
	for _, item := range feed.Items {
		pub, err := parseRSSTime(item.Published)
		if err != nil {
			errs = append(errs, fmt.Errorf("error parsing time, still processing episode: %w", err))
			pub = time.Time{}
		}

		desc := item.Description
		if desc == "" && item.ITunesExt.Summary != "" {
			desc = item.ITunesExt.Summary
		}

		if item.Enclosures == nil ||
			len(item.Enclosures) == 0 ||
			item.Enclosures[0] == nil ||
			item.Enclosures[0].URL == "" {
			errs = append(errs, fmt.Errorf("could not read download URL, skipping episode '%s'", episodeGUID(item)))
			continue
		}

		if _, ok := MimeToExt[item.Enclosures[0].Type]; !ok {
			errs = append(errs, fmt.Errorf(
				"unsupported file type '%s', skipping episode '%s'",
				item.Enclosures[0].Type,
				episodeGUID(item),
			))
			continue
		}

		episode := Episode{
			GUID:         episodeGUID(item),
			Title:        truncate(item.Title, 1000),
			Description:  truncate(desc, 10000),
			DownloadURL:  item.Enclosures[0].URL,
			MimeType:     item.Enclosures[0].Type,
			DurationSecs: parseDuration(item),
			PublishedAt:  pub,
		}

		episodes = append(episodes, episode)
	}

	return episodes, errs
}

func feedGUID(feed *gofeed.Feed) string {
	// use guid if defined
	if feed.Extensions != nil &&
		feed.Extensions["podcast"] != nil &&
		feed.Extensions["podcast"]["guid"] != nil &&
		len(feed.Extensions["podcast"]["guid"]) > 0 &&
		feed.Extensions["podcast"]["guid"][0].Value != "" {
		return feed.Extensions["podcast"]["guid"][0].Value
	}

	// fallback to hash of feed link or title otherwise
	hashIn := feed.FeedLink
	if hashIn == "" {
		hashIn = feed.Title
	}

	h := sha256.New()
	_, _ = h.Write([]byte(hashIn))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func episodeGUID(feedItem *gofeed.Item) string {
	if feedItem.GUID != "" {
		return feedItem.GUID
	}

	// fallback to title + pub date if guid not present
	hashIn := feedItem.Title + feedItem.Published
	h := sha256.New()
	_, _ = h.Write([]byte(hashIn))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func truncate(s string, l int) string {
	if len(s) > l {
		return s[:l-1]
	}
	return s
}

func parseRSSTime(s string) (time.Time, error) {
	formats := []string{time.RFC822, time.RFC822Z, time.RFC1123, time.RFC1123Z}
	var t time.Time
	var err error
	for _, f := range formats {
		t, err = time.Parse(f, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("failed to parse time '%s'", s)
}

func parseDuration(item *gofeed.Item) int {
	if item.ITunesExt == nil || item.ITunesExt.Duration == "" {
		return 0
	}

	durStr := item.ITunesExt.Duration
	colonCount := strings.Count(durStr, ":")

	if colonCount == 0 {
		duration, _ := strconv.Atoi(durStr)
		return duration
	}

	if colonCount > 2 {
		return 0
	}

	format := "15:04:05"
	if colonCount == 1 {
		format = "04:05"
	}

	t, err := time.Parse(format, durStr)
	if err != nil {
		return 0
	}

	// use 0000 origin instead of the default time.Time{} 0001
	dur := t.Sub(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))
	return int(math.Round(dur.Seconds()))
}
