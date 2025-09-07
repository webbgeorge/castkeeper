package podcasts

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/webbgeorge/castkeeper/pkg/util"
	"github.com/webbgeorge/gopodcast"
	"gorm.io/gorm"
)

type ParseErrors struct {
	parseErrs []error
}

func (pe ParseErrors) Error() string {
	errStrs := make([]string, 0)
	for _, err := range pe.parseErrs {
		errStrs = append(errStrs, err.Error())
	}
	return fmt.Sprintf("%d errors whilst parsing episodes: %s", len(errStrs), strings.Join(errStrs, ", "))
}

func (pe ParseErrors) Is(target error) bool {
	_, ok := target.(ParseErrors)
	return ok
}

type FeedService struct {
	HTTPClient *http.Client
}

func (s *FeedService) ParseFeed(ctx context.Context, feedURL string, creds *PodcastCredentials) (Podcast, []Episode, error) {
	err := util.ValidateExtURL(feedURL)
	if err != nil {
		return Podcast{}, nil, fmt.Errorf("invalid feedURL '%s': %w", feedURL, err)
	}

	fp := gopodcast.NewParser()
	fp.HTTPClient = s.HTTPClient

	if creds != nil {
		fp.AuthCredentials = &gopodcast.AuthCredentials{
			Username: creds.Username,
			Password: creds.Password,
		}
	}

	feed, err := fp.ParseFeedFromURL(ctx, feedURL)
	if err != nil {
		return Podcast{}, nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	return podcastFromFeed(feedURL, feed)
}

func podcastFromFeed(feedURL string, feed *gopodcast.Podcast) (Podcast, []Episode, error) {
	guid := feedGUID(feed)

	var lastEpisodeAt *time.Time
	episodes, err := episodesFromFeed(feed, guid)
	if len(episodes) > 0 {
		// feed items are sorted oldest to newest
		lastEpisodeAt = &episodes[len(episodes)-1].PublishedAt
	}

	author := "unknown"
	if feed.ITunesAuthor != "" {
		author = feed.ITunesAuthor
	}

	podcast := Podcast{
		GUID:          guid,
		Title:         truncate(feed.Title, 500),
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

	return podcast, episodes, err
}

func episodesFromFeed(feed *gopodcast.Podcast, podcastGUID string) ([]Episode, error) {
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

		mimeType, err := DetectMIMEType(item.Enclosure)
		if err != nil {
			errs = append(errs, fmt.Errorf(
				"failed to parse episode '%s', skipping: %w",
				episodeGUID(item),
				err,
			))
			continue
		}

		episode := Episode{
			GUID:         episodeGUID(item),
			PodcastGUID:  podcastGUID,
			Title:        truncate(item.Title, 500),
			Description:  truncate(desc, 10000),
			DownloadURL:  item.Enclosure.URL,
			MimeType:     mimeType,
			DurationSecs: parseDuration(item),
			PublishedAt:  pub,
		}

		episodes = append(episodes, episode)
	}

	slices.SortFunc(episodes, func(a, b Episode) int {
		return a.PublishedAt.Compare(b.PublishedAt)
	})

	var err error
	if len(errs) > 0 {
		err = ParseErrors{parseErrs: errs}
	}

	return episodes, err
}

func feedGUID(feed *gopodcast.Podcast) string {
	if feed.PodcastGUID != "" {
		return uuid.NewV5(uuid.NamespaceOID, feed.PodcastGUID).String()
	}

	// fallback to hash of feed link or title otherwise
	hashIn := feed.AtomLink.Href
	if hashIn == "" {
		hashIn = feed.Link
	}
	if hashIn == "" {
		hashIn = feed.Title
	}

	return uuid.NewV5(uuid.NamespaceOID, hashIn).String()
}

func episodeGUID(feedItem *gopodcast.Item) string {
	if feedItem.GUID.Text != "" {
		return uuid.NewV5(uuid.NamespaceOID, feedItem.GUID.Text).String()
	}

	// fallback to title + pub date or Enclosure URL if guid not present
	guidSuffix := ""
	if feedItem.PubDate != nil {
		guidSuffix = time.Time(*feedItem.PubDate).Format(time.RFC3339)
	}
	if guidSuffix == "" {
		guidSuffix = feedItem.Enclosure.URL
	}

	return uuid.NewV5(uuid.NamespaceOID, feedItem.Title+guidSuffix).String()
}

func truncate(s string, l int) string {
	if len(s) > l {
		return s[:l]
	}
	return s
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

func feedCategories(feed *gopodcast.Podcast) []Category {
	cats := make([]Category, 0)
	for _, c := range feed.ITunesCategory {
		if c.Text != "" {
			cat := Category{
				Name: c.Text,
			}
			if c.SubCategory != nil && c.SubCategory.Text != "" {
				cat.SubCategory = &Category{
					Name: c.SubCategory.Text,
				}
			}
			cats = append(cats, cat)
		}
	}
	return cats
}

func GenerateFeed(ctx context.Context, baseURL string, db *gorm.DB, podcastGuid string) (*gopodcast.Podcast, error) {
	pod, err := GetPodcast(ctx, db, podcastGuid)
	if err != nil {
		return nil, err
	}

	eps, err := ListEpisodes(ctx, db, pod.GUID)
	if err != nil {
		return nil, err
	}

	return feedFromPodcast(baseURL, pod, eps)
}

func feedFromPodcast(baseURL string, pod Podcast, eps []Episode) (*gopodcast.Podcast, error) {
	categories := make([]gopodcast.ITunesCategory, 0)
	for _, cat := range pod.Categories {
		if cat.Name == "" {
			continue
		}
		iCat := gopodcast.ITunesCategory{
			Text: cat.Name,
		}
		if cat.SubCategory != nil && cat.SubCategory.Name != "" {
			iCat.SubCategory = &gopodcast.ITunesCategory{
				Text: cat.SubCategory.Name,
			}
		}
		categories = append(categories, iCat)
	}

	feed := &gopodcast.Podcast{
		AtomLink: gopodcast.AtomLink{
			Href: fmt.Sprintf("%s/feeds/%s", baseURL, pod.GUID),
			Rel:  "self",
			Type: "application/rss+xml",
		},
		Title:          pod.Title,
		Description:    gopodcast.Description{Text: pod.Description},
		Link:           pod.Link,
		Language:       pod.Language,
		ITunesCategory: categories,
		ITunesExplicit: gopodcast.Bool(pod.IsExplicit),
		ITunesImage:    gopodcast.ITunesImage{Href: fmt.Sprintf("%s/podcasts/%s/image", baseURL, pod.GUID)},
	}

	for _, ep := range eps {
		if ep.Status != EpisodeStatusSuccess {
			continue
		}

		pubDate := gopodcast.Time(ep.PublishedAt)
		feed.Items = append(feed.Items, &gopodcast.Item{
			Title:       ep.Title,
			Description: &gopodcast.Description{Text: ep.Description},
			Enclosure: gopodcast.Enclosure{
				Length: ep.Bytes,
				Type:   ep.MimeType,
				URL:    fmt.Sprintf("%s/episodes/%s/download", baseURL, ep.GUID),
			},
			GUID:    gopodcast.ItemGUID{Text: ep.GUID},
			PubDate: &pubDate,
		})
	}

	return feed, nil
}
