package podcasts

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/util"
	"github.com/webbgeorge/gopodcast"
)

type FeedService struct {
	HTTPClient *http.Client
}

func (s *FeedService) ParseFeed(ctx context.Context, feedURL string) (*gopodcast.Podcast, error) {
	err := util.ValidateExtURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid feedURL '%s': %w", feedURL, err)
	}

	fp := gopodcast.NewParser()
	fp.HTTPClient = s.HTTPClient

	feed, err := fp.ParseFeedFromURL(ctx, feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	return feed, nil
}

func PodcastFromFeed(feedURL string, feed *gopodcast.Podcast) Podcast {
	var lastEpisodeAt *time.Time
	episodes, _ := EpisodesFromFeed(feed)
	if len(episodes) > 0 {
		// feed items are sorted oldest to newest
		lastEpisodeAt = &episodes[len(episodes)-1].PublishedAt
	}

	author := "unknown"
	if feed.ITunesAuthor != "" {
		author = feed.ITunesAuthor
	}

	podcast := Podcast{
		GUID:          feedGUID(feed),
		Title:         truncate(feed.Title, 1000),
		Author:        author,
		Description:   truncate(feed.Description.Text, 10000),
		Language:      feed.Language,
		Link:          feed.Link,
		Categories:    feedCategories(feed),
		IsExplicit:    bool(feed.ITunesExplicit),
		ImageURL:      feed.ITunesImage.Href,
		FeedURL:       feedURL,
		LastCheckedAt: nil,
		LastEpisodeAt: lastEpisodeAt,
	}

	return podcast
}

func EpisodesFromFeed(feed *gopodcast.Podcast) ([]Episode, []error) {
	episodes := make([]Episode, 0)
	errs := make([]error, 0)
	for _, item := range feed.Items {
		desc := ""
		if item.Description != nil {
			desc = item.Description.Text
		}

		pub := time.Time{}
		if item.PubDate != nil {
			pub = time.Time(*item.PubDate)
		}

		if item.Enclosure.URL == "" {
			errs = append(errs, fmt.Errorf("could not read download URL, skipping episode '%s'", episodeGUID(item)))
			continue
		}

		if _, ok := MimeToExt[item.Enclosure.Type]; !ok {
			errs = append(errs, fmt.Errorf(
				"unsupported file type '%s', skipping episode '%s'",
				item.Enclosure.Type,
				episodeGUID(item),
			))
			continue
		}

		episode := Episode{
			GUID:         episodeGUID(item),
			Title:        truncate(item.Title, 1000),
			Description:  truncate(desc, 10000),
			DownloadURL:  item.Enclosure.URL,
			MimeType:     item.Enclosure.Type,
			DurationSecs: parseDuration(item),
			PublishedAt:  pub,
		}

		episodes = append(episodes, episode)
	}

	slices.SortFunc(episodes, func(a, b Episode) int {
		return a.PublishedAt.Compare(b.PublishedAt)
	})

	return episodes, errs
}

// TODO I'm not confident this method of getting a unique ID is sufficient
func feedGUID(feed *gopodcast.Podcast) string {
	if feed.PodcastGUID != "" {
		return util.SanitiseGUID(feed.PodcastGUID)
	}

	// fallback to hash of feed link or title otherwise
	hashIn := feed.AtomLink.Href
	if hashIn == "" {
		hashIn = feed.Link
	}
	if hashIn == "" {
		hashIn = feed.Title
	}

	h := sha256.New()
	_, _ = h.Write([]byte(hashIn))
	newGUID := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return util.SanitiseGUID(newGUID)
}

// TODO I'm not confident this method of getting a unique ID is sufficient
func episodeGUID(feedItem *gopodcast.Item) string {
	if feedItem.GUID.Text != "" {
		return util.SanitiseGUID(feedItem.GUID.Text)
	}

	// fallback to title + pub date or Enclosure URL if guid not present

	guidSuffix := ""
	if feedItem.PubDate != nil {
		guidSuffix = time.Time(*feedItem.PubDate).Format(time.RFC3339)
	}
	if guidSuffix == "" {
		guidSuffix = feedItem.Enclosure.URL
	}

	hashIn := feedItem.Title + guidSuffix
	h := sha256.New()
	_, _ = h.Write([]byte(hashIn))
	newGUID := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return util.SanitiseGUID(newGUID)
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

func parseDuration(item *gopodcast.Item) int {
	if item.ITunesDuration == "" {
		return 0
	}

	durStr := item.ITunesDuration
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

func feedCategories(feed *gopodcast.Podcast) []string {
	cats := make([]string, 0)
	for _, c := range feed.ITunesCategory {
		if c.Text != "" {
			cats = append(cats, c.Text)
			if c.SubCategory != nil && c.SubCategory.Text != "" {
				cats = append(cats, fmt.Sprintf("%s:%s", c.Text, c.SubCategory.Text))
			}
		}
	}
	return cats
}
