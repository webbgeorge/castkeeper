package podcasts

import (
	"context"
	"fmt"

	"github.com/webbgeorge/gopodcast"
	"gorm.io/gorm"
)

func GenerateFeed(ctx context.Context, baseURL string, db *gorm.DB, podcastGuid string) (*gopodcast.Podcast, error) {
	pod, err := GetPodcast(ctx, db, podcastGuid)
	if err != nil {
		// TODO handle not found error
		return nil, err
	}

	eps, err := ListEpisodes(ctx, db, pod.GUID)
	if err != nil {
		return nil, err
	}

	categories := make([]gopodcast.ITunesCategory, 0)
	for _, cat := range pod.Categories {
		categories = append(categories, gopodcast.ITunesCategory{Text: cat})
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
