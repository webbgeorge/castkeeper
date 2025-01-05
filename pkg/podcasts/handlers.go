package podcasts

import (
	"context"
	"errors"
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

		// TODO use gorilla form to get from a struct
		feedURL := r.PostFormValue("feedUrl")
		feed, err := ParseFeed(ctx, feedURL)
		if err != nil {
			return framework.Render(ctx, w, 200, pages.Subscribe(true, err))
		}
		podcast := PodcastFromFeed(feedURL, feed)

		if err = db.Create(&podcast).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				// TODO better error handling in view (send string instead?)
				err2 := errors.New("already subscribed to this feed")
				return framework.Render(ctx, w, 200, pages.Subscribe(true, err2))
			}
			return err
		}

		return framework.Render(ctx, w, 200, pages.Subscribe(true, nil))
	}
}
