package migrations

import (
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
	"gorm.io/gorm"
)

type Migration002AddEncryptedValue struct{}

func (m Migration002AddEncryptedValue) Name() string {
	return "002-add-encrypted-value"
}

func (m Migration002AddEncryptedValue) Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&encryption.EncryptedValue{}); err != nil {
		return err
	}
	return nil
}
