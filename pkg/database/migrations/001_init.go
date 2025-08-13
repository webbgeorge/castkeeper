package migrations

import (
	"github.com/webbgeorge/castkeeper/pkg/auth/sessions"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

type Migration001Init struct{}

func (m Migration001Init) Name() string {
	return "001-init"
}

func (m Migration001Init) Migrate(db *gorm.DB) error {
	// AutoMigrate initially as the DB already exists for some users
	if err := db.AutoMigrate(&podcasts.Podcast{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&podcasts.Episode{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&users.User{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&sessions.Session{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&framework.QueueTask{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&framework.ScheduledTaskState{}); err != nil {
		return err
	}

	// if users exist at this point, they should be made Admins
	// as this initial migration was introduced to support the
	// access control feature - prior to this, users had
	// permission to do everything. Any users created after this
	// migration being run will be explicitly assigned an access
	// level.
	var userList []users.User
	result := db.
		Order("id asc").
		Find(&userList)
	if result.Error != nil {
		return result.Error
	}
	for _, user := range userList {
		user.AccessLevel = users.AccessLevelAdmin
		if err := db.Save(&user).Error; err != nil {
			return err
		}
	}

	return nil
}
