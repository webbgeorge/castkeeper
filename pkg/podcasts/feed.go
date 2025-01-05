package podcasts

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
)

func ParseFeed(ctx context.Context, feedURL string) (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
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
	episodes := EpisodesFromFeed(feed)
	if len(episodes) > 0 {
		// feed items are sorted oldest to newest
		lastEpisodeAt = &episodes[len(episodes)-1].PublishedAt
	}

	podcast := Podcast{
		Title:         truncate(feed.Title, 1000),
		FeedURL:       feedURL,
		GUID:          feedGUID(feed),
		LastCheckedAt: nil,
		LastEpisodeAt: lastEpisodeAt,
	}

	return podcast
}

func EpisodesFromFeed(feed *gofeed.Feed) []Episode {
	episodes := make([]Episode, 0)
	for _, item := range feed.Items {
		pub, err := parseRSSTime(item.Published)
		if err != nil {
			// TODO log warning
			fmt.Printf("error parsing time: %s", err) // TODO logger
			pub = time.Time{}
		}

		desc := item.Description
		if desc == "" && item.ITunesExt.Summary != "" {
			desc = item.ITunesExt.Summary
		}

		episode := Episode{
			Title:       truncate(item.Title, 1000),
			Description: truncate(desc, 10000),
			GUID:        episodeGUID(item),
			PublishedAt: pub,
		}

		episodes = append(episodes, episode)
	}

	return episodes
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
	return string(h.Sum(nil))
}

func episodeGUID(feedItem *gofeed.Item) string {
	if feedItem.GUID != "" {
		return feedItem.GUID
	}

	// fallback to title + pub date if guid not present
	hashIn := feedItem.Title + feedItem.Published
	h := sha256.New()
	_, _ = h.Write([]byte(hashIn))
	return string(h.Sum(nil))
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
