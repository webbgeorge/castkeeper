package podcasts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

const (
	EpisodeStatusPending = "pending"
	EpisodeStatusSuccess = "success"
	EpisodeStatusFailed  = "failed"
)

type Podcast struct {
	gorm.Model
	GUID          string `gorm:"primaryKey" validate:"required,gte=1,lte=1000"`
	Title         string `validate:"required,gte=1,lte=1000"`
	FeedURL       string `validate:"required,http_url,lte=1000"`
	LastCheckedAt *time.Time
	LastEpisodeAt *time.Time
}

type Episode struct {
	gorm.Model
	GUID         string  `gorm:"primaryKey" validate:"required,gte=1,lte=1000"`
	PodcastGUID  string  `validate:"required"`
	Podcast      Podcast `validate:"-" gorm:"foreignKey:PodcastGUID"`
	Title        string  `validate:"required,gte=1,lte=1000"`
	Description  string  `validate:"lte=10000"`
	DownloadURL  string  `validate:"required,http_url,lte=1000"`
	MimeType     string  `validate:"required,oneof=audio/mpeg audio/x-m4a video/mp4 video/quicktime"`
	PublishedAt  time.Time
	Status       string `validate:"required,oneof=pending failed success"`
	FailureCount int    `validate:"gte=0"`
	LastFailedAt *time.Time
}

var MimeToExt = map[string]string{
	"audio/mpeg":      "mp3",
	"audio/x-m4a":     "m4a",
	"video/mp4":       "mp4",
	"video/quicktime": "mov",
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func (p *Podcast) BeforeSave(tx *gorm.DB) error {
	err := validate.Struct(p)
	if err != nil {
		return fmt.Errorf("podcast not valid: %w", err)
	}
	return nil
}

func (e *Episode) BeforeSave(tx *gorm.DB) error {
	err := validate.Struct(e)
	if err != nil {
		return fmt.Errorf("episode not valid: %w", err)
	}
	return nil
}

func ListPodcasts(ctx context.Context, db *gorm.DB) ([]Podcast, error) {
	var podcasts []Podcast
	result := db.Find(&podcasts)
	if result.Error != nil {
		return nil, result.Error
	}
	return podcasts, nil
}

func ListEpisodes(ctx context.Context, db *gorm.DB) ([]Episode, error) {
	var episodes []Episode
	result := db.Find(&episodes)
	if result.Error != nil {
		return nil, result.Error
	}
	return episodes, nil
}

func UpdatePodcastTimes(ctx context.Context, db *gorm.DB, podcast *Podcast, lastCheckedAt, lastEpisodeAt *time.Time) error {
	result := db.
		Model(podcast).
		Select("LastCheckedAt", "LastEpisodeAt").
		Updates(Podcast{LastCheckedAt: lastCheckedAt, LastEpisodeAt: lastEpisodeAt})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

var ErrEpisodeNotFound = errors.New("episode not found")

func GetPendingEpisode(ctx context.Context, db *gorm.DB, hasNotFailedSince time.Time) (Episode, error) {
	var episode Episode
	result := db.First(
		&episode,
		"status = ?", EpisodeStatusPending,
		"last_failed_at < ?", hasNotFailedSince,
	)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return episode, ErrEpisodeNotFound
		}
		return episode, result.Error
	}
	return episode, nil
}

func GetPodcast(ctx context.Context, db *gorm.DB, guid string) (Podcast, error) {
	var podcast Podcast
	result := db.First(&podcast, guid)
	if result.Error != nil {
		return podcast, result.Error
	}
	return podcast, nil
}

func UpdateEpisodeStatus(ctx context.Context, db *gorm.DB, episode *Episode, status string) error {
	result := db.
		Model(episode).
		Select("Status").
		Updates(Episode{Status: status})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func UpdateEpisodeFailureCount(ctx context.Context, db *gorm.DB, episode *Episode, failureCount int) error {
	now := time.Now()
	result := db.
		Model(episode).
		Select("FailureCount", "LastFailedAt").
		Updates(Episode{FailureCount: failureCount, LastFailedAt: &now})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
