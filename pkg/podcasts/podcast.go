package podcasts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"gorm.io/gorm"
)

const (
	EpisodeStatusPending = "pending"
	EpisodeStatusSuccess = "success"
	EpisodeStatusFailed  = "failed"
)

type Podcast struct {
	GUID          string     `gorm:"primaryKey" validate:"required,gte=1,lte=1000"`
	Title         string     `validate:"required,gte=1,lte=1000"`
	Author        string     `validate:"required,gte=1,lte=1000"`
	Description   string     `validate:"lte=10000"`
	Language      string     `validate:"lte=10"`
	Link          string     `validate:"lte=1000"`
	Categories    []Category `gorm:"serializer:json" validate:"lte=25"`
	IsExplicit    bool
	ImageURL      string `validate:"lte=1000"`
	FeedURL       string `validate:"required,http_url,lte=1000"`
	LastCheckedAt *time.Time
	LastEpisodeAt *time.Time
	Credentials   *encryption.EncryptedValue `validate:"-" gorm:"embedded"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type Category struct {
	Name        string `validate:"required,gte=1,lte=100"`
	SubCategory *Category
}

type PodcastCredentials struct {
	Username string `validate:"lte=256"`
	Password string `validate:"lte=256"`
}

type Episode struct {
	GUID         string  `gorm:"primaryKey" validate:"required,gte=1,lte=1000"`
	PodcastGUID  string  `validate:"required"`
	Podcast      Podcast `validate:"-" gorm:"foreignKey:PodcastGUID"`
	Title        string  `validate:"required,gte=1,lte=1000"`
	Description  string  `validate:"lte=10000"`
	DownloadURL  string  `validate:"required,http_url,lte=1000"`
	Bytes        int64
	MimeType     string `validate:"required,oneof=audio/mpeg audio/x-m4a video/mp4 video/quicktime"`
	DurationSecs int    `validate:"gte=0"`
	PublishedAt  time.Time
	Status       string `validate:"required,oneof=pending failed success"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
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

func (pc PodcastCredentials) Validate() error {
	err := validate.Struct(pc)
	if err != nil {
		return fmt.Errorf("podcast credentials not valid: %w", err)
	}
	return nil
}

func AddPodcast(
	ctx context.Context,
	db *gorm.DB,
	feedService *FeedService,
	encService *encryption.EncryptedValueService,
	feedURL string,
	creds *PodcastCredentials,
) (Podcast, error) {
	podcast, _, err := feedService.ParseFeed(ctx, feedURL, creds)
	if err != nil {
		if !errors.Is(err, ParseErrors{}) {
			framework.GetLogger(ctx).ErrorContext(ctx, "error parsing feed", "error", err)
			return podcast, err
		}
		framework.GetLogger(ctx).WarnContext(ctx, fmt.Sprintf("some episodes of podcast '%s' had parsing errors: %s", podcast.GUID, err.Error()))
		// continue even with some episode parse failures...
	}

	if creds != nil {
		if err := creds.Validate(); err != nil {
			return podcast, err
		}

		credsData, err := json.Marshal(creds)
		if err != nil {
			return podcast, err
		}

		ev, err := encService.Encrypt(
			credsData,
			[]byte(podcast.FeedURL),
		)
		if err != nil {
			return podcast, err
		}
		podcast.Credentials = &ev
	}

	if err = db.Create(&podcast).Error; err != nil {
		return podcast, err
	}

	return podcast, nil
}

func GetCredentials(encService *encryption.EncryptedValueService, podcast Podcast) (*PodcastCredentials, error) {
	if podcast.Credentials == nil {
		return nil, nil
	}

	data, err := encService.Decrypt(*podcast.Credentials, []byte(podcast.FeedURL))
	if err != nil {
		return nil, err
	}

	var creds PodcastCredentials
	err = json.Unmarshal(data, &creds)
	if err != nil {
		return nil, err
	}

	return &creds, nil
}

func ListPodcasts(ctx context.Context, db *gorm.DB) ([]Podcast, error) {
	var podcasts []Podcast
	result := db.Find(&podcasts)
	if result.Error != nil {
		return nil, result.Error
	}
	return podcasts, nil
}

func ListEpisodes(ctx context.Context, db *gorm.DB, podcastGUID string) ([]Episode, error) {
	var episodes []Episode
	result := db.
		Order("published_at desc").
		Find(&episodes, "podcast_guid = ?", podcastGUID)
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

func GetPodcast(ctx context.Context, db *gorm.DB, guid string) (Podcast, error) {
	var podcast Podcast
	result := db.First(&podcast, "guid = ?", guid)
	if result.Error != nil {
		return podcast, result.Error
	}
	return podcast, nil
}

func GetEpisode(ctx context.Context, db *gorm.DB, guid string) (Episode, error) {
	var episode Episode
	result := db.Preload("Podcast").First(&episode, "guid = ?", guid)
	if result.Error != nil {
		return episode, result.Error
	}
	return episode, nil
}

func UpdateEpisodeStatus(ctx context.Context, db *gorm.DB, episode *Episode, status string, fileBytes *int64) error {
	fields := []string{"Status"}
	epUpdate := Episode{Status: status}
	if fileBytes != nil {
		fields = append(fields, "Bytes")
		epUpdate.Bytes = *fileBytes
	}

	result := db.
		Model(episode).
		Select(fields).
		Updates(epUpdate)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
