package migrations

import (
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

type Migration002AddPodcastCredentials struct{}

func (m Migration002AddPodcastCredentials) Name() string {
	return "002-add-podcast-credentials"
}

func (m Migration002AddPodcastCredentials) Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&podcasts.Podcast{}); err != nil {
		return err
	}
	return nil
}
