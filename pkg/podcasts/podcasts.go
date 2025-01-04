package podcasts

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Podcast struct {
	gorm.Model
	Name    string `validate:"required,gte=1,lte=1000"`
	FeedURL string `validate:"required,http_url"`
}

var validate = validator.New(validator.WithRequiredStructEnabled())

func (p *Podcast) BeforeSave(tx *gorm.DB) error {
	err := validate.Struct(p)
	if err != nil {
		return fmt.Errorf("podcast not valid: %w", err)
	}
	return nil
}
