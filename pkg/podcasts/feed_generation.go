package podcasts

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Parser struct {
	HTTPClient *http.Client
}

func (p *Parser) ParseFeedFromURL(ctx context.Context, url string) (*Feed, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	res, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("non-200 http response '%d'", res.StatusCode)
	}

	var feed Feed
	err = xml.NewDecoder(res.Body).Decode(&feed)
	if err != nil {
		return nil, err
	}

	return &feed, nil
}

// TODO better types on XML structs, plus custom marshallers

type Feed struct {
	XMLName      xml.Name `xml:"rss"`
	Version      string   `xml:"version,attr"`
	XMLNSContent string   `xml:"xmlns:content,attr"`
	XMLNSPodcast string   `xml:"xmlns:podcast,attr"`
	XMLNSAtom    string   `xml:"xmlns:atom,attr"`
	XMLNSITunes  string   `xml:"xmlns:itunes,attr"`
	Channel      *Channel
}

func (f *Feed) WriteFeedXML(w io.Writer) error {
	w.Write([]byte(xml.Header))
	return xml.NewEncoder(w).Encode(f)
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`

	// PSP Required
	AtomLink       AtomLink
	Title          string `xml:"title"`
	Description    Description
	Link           string `xml:"link"`
	Language       string `xml:"language"`
	ITunesCategory []ITunesCategory
	ITunesExplicit bool `xml:"itunes:explicit"`
	ITunesImage    ITunesImage

	// PSP Recommended
	PodcastLocked string `xml:"podcast:locked,omitempty"`
	PodcastGUID   string `xml:"podcast:guid,omitempty"`
	ITunesAuthor  string `xml:"itunes:author,omitempty"`

	// PSP Optional
	Copyright      string          `xml:"copyright,omitempty"`
	PodcastText    *PodcastText    `xml:",omitempty"`
	PodcastFunding *PodcastFunding `xml:",omitempty"`
	ITunesType     string          `xml:"itunes:type,omitempty"`
	ITunesComplete string          `xml:"itunes:complete,omitempty"`

	Items []*Item
}

type AtomLink struct {
	XMLName xml.Name `xml:"atom:link"`
	Href    string   `xml:"href,attr"`
	Rel     string   `xml:"rel,attr"`
	Type    string   `xml:"type,attr"`
}

type Description struct {
	XMLName xml.Name `xml:"description"`
	Text    string   `xml:",cdata"`
}

type ITunesCategory struct {
	XMLName        xml.Name `xml:"itunes:category"`
	Text           string   `xml:"text,attr"`
	ITunesCategory []ITunesCategory
}

type ITunesImage struct {
	XMLName xml.Name `xml:"itunes:image"`
	Href    string   `xml:"href,attr"`
}

type PodcastText struct {
	XMLName xml.Name `xml:"podcast:txt"`
	Purpose string   `xml:"purpose,attr,omitempty"`
	Text    string   `xml:",chardata"`
}

type PodcastFunding struct {
	XMLName xml.Name `xml:"podcast:funding"`
	URL     string   `xml:"url,attr"`
	Text    string   `xml:",chardata"`
}

type Item struct {
	XMLName xml.Name `xml:"item"`

	// PSP required
	Title     string `xml:"title"`
	Enclosure Enclosure
	GUID      ItemGUID

	// PSP Recommended
	Link              string              `xml:"link,omitempty"`
	PubDate           string              `xml:"pubDate,omitempty"` // TODO time.Time with custom marshaller
	Description       *Description        `xml:",omitempty"`
	ITunesDuration    string              `xml:"itunes:duration,omitempty"`
	ITunesImage       *ITunesImage        `xml:",omitempty"`
	ITunesExplicit    *bool               `xml:"itunes:explicit,omitempty"`
	PodcastTranscript []PodcastTranscript `xml:",omitempty"`

	// PSP Optional
	ITunesEpisode     string `xml:"itunes:episode,omitempty"`
	ITunesSeason      string `xml:"itunes:season,omitempty"`
	ITunesEpisodeType string `xml:"itunes:episodeType,omitempty"`
	ITunesBlock       string `xml:"itunes:block,omitempty"`
}

type Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	Length  int64    `xml:"length,attr"`
	Type    string   `xml:"type,attr"`
	URL     string   `xml:"url,attr"`
}

type ItemGUID struct {
	XMLName     xml.Name `xml:"guid"`
	IsPermaLink *bool    `xml:"isPermaLink,attr,omitempty"`
	Text        string   `xml:",chardata"`
}

type PodcastTranscript struct {
	XMLName  xml.Name `xml:"podcast:transcript"`
	URL      string   `xml:"url,attr"`
	Type     string   `xml:"type,attr"`
	Rel      string   `xml:"rel,attr,omitempty"`
	Language string   `xml:"language,attr,omitempty"`
}

func GenerateFeed(ctx context.Context, baseURL string, db *gorm.DB, podcastGuid string) (*Feed, error) {
	pod, err := GetPodcast(ctx, db, podcastGuid)
	if err != nil {
		// TODO handle not found error
		return nil, err
	}

	eps, err := ListEpisodes(ctx, db, pod.GUID)
	if err != nil {
		return nil, err
	}

	categories := make([]ITunesCategory, 0)
	for _, cat := range pod.Categories {
		categories = append(categories, ITunesCategory{Text: cat})
	}

	channel := &Channel{
		AtomLink: AtomLink{
			Href: fmt.Sprintf("%s/feeds/%s", baseURL, pod.GUID),
			Rel:  "self",
			Type: "application/rss+xml",
		},
		Title:          pod.Title,
		Description:    Description{Text: pod.Description},
		Link:           pod.Link,
		Language:       pod.Language,
		ITunesCategory: categories,
		ITunesExplicit: pod.IsExplicit,
		ITunesImage:    ITunesImage{Href: fmt.Sprintf("%s/podcasts/%s/image", baseURL, pod.GUID)},
	}

	for _, ep := range eps {
		if ep.Status != EpisodeStatusSuccess {
			continue
		}

		channel.Items = append(channel.Items, &Item{
			Title:       ep.Title,
			Description: &Description{Text: ep.Description},
			Enclosure: Enclosure{
				Length: ep.Bytes,
				Type:   ep.MimeType,
				URL:    fmt.Sprintf("%s/episodes/%s/download", baseURL, ep.GUID),
			},
			GUID:    ItemGUID{Text: ep.GUID},
			PubDate: ep.PublishedAt.Format(time.RFC1123),
		})
	}

	feed := &Feed{
		Version:      "2.0",
		XMLNSContent: "http://purl.org/rss/1.0/modules/content/",
		XMLNSPodcast: "https://podcastindex.org/namespace/1.0",
		XMLNSAtom:    "http://www.w3.org/2005/Atom",
		XMLNSITunes:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel:      channel,
	}

	return feed, nil
}
