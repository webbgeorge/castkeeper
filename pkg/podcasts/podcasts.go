package podcasts

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"
)

type Podcast struct {
	gorm.Model
	Name    string `validate:"required,gte=1,lte=1000"`
	FeedURL string `validate:"required,http_url"`
	GUID    string `gorm:"uniqueIndex" validate:"required,gte=1,lte=1000"`
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func (p *Podcast) BeforeSave(tx *gorm.DB) error {
	err := validate.Struct(p)
	if err != nil {
		return fmt.Errorf("podcast not valid: %w", err)
	}
	return nil
}

func PodcastFromFeed(ctx context.Context, feedURL string) (Podcast, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return Podcast{}, fmt.Errorf("failed to parse feed: %w", err)
	}

	if feed.ITunesExt == nil {
		return Podcast{}, errors.New("feed is not a podcast")
	}

	guid, err := fallbackGUID(feedURL)
	if err != nil {
		return Podcast{}, err
	}

	if feed.Extensions != nil &&
		feed.Extensions["podcast"] != nil &&
		feed.Extensions["podcast"]["guid"] != nil &&
		len(feed.Extensions["podcast"]["guid"]) > 0 &&
		feed.Extensions["podcast"]["guid"][0].Value != "" {
		guid = feed.Extensions["podcast"]["guid"][0].Value
	}

	podcast := Podcast{
		Name:    feed.Title,
		FeedURL: feedURL,
		GUID:    guid,
	}

	return podcast, nil
}

func fallbackGUID(feedURL string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(feedURL))
	if err != nil {
		return "", err
	}
	return string(h.Sum(nil)), nil
}
