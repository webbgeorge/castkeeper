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
	if db.Migrator().HasColumn(&podcasts.Podcast{}, "EncryptedData") {
		return nil
	}
	if err := db.Migrator().AddColumn(&podcasts.Podcast{}, "EncryptedData"); err != nil {
		return err
	}
	if err := db.Migrator().AddColumn(&podcasts.Podcast{}, "KeyVersion"); err != nil {
		return err
	}
	if err := db.Migrator().AddColumn(&podcasts.Podcast{}, "Salt"); err != nil {
		return err
	}
	return nil
}
