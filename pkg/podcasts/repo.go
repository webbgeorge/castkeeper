package podcasts

import (
	"context"

	"gorm.io/gorm"
)

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
