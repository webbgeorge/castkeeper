package podcasts

import (
	"context"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"gorm.io/gorm"
)

func NewSubscribeGetHandler() framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return framework.Render(ctx, w, 200, pages.Subscribe(false, nil))
	}
}

func NewSubscribePostHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		err := r.ParseForm()
		if err != nil {
			return framework.Render(ctx, w, 200, pages.Subscribe(true, err))
		}

		// TODO get feedurl from request, fetch the feed and verify it is a podcast

		podcast := Podcast{
			// TODO get from feed data
			Name: "TODO",

			// TODO use gorilla form to get from a struct
			FeedURL: r.PostFormValue("feedUrl"),
		}

		if err = db.Create(&podcast).Error; err != nil {
			return err
		}

		return framework.Render(ctx, w, 200, pages.Subscribe(true, nil))
	}
}
