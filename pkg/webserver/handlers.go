package webserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"gorm.io/gorm"
)

func NewHomeHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path != "/" {
			// handle fallback on home route
			return framework.HttpNotFound()
		}

		pods, err := podcasts.ListPodcasts(ctx, db)
		if err != nil {
			return err
		}
		return framework.Render(ctx, w, 200, pages.Home(pods))
	}
}

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
		feed, err := podcasts.ParseFeed(ctx, feedURL)
		if err != nil {
			return framework.Render(ctx, w, 200, pages.Subscribe(true, err))
		}
		podcast := podcasts.PodcastFromFeed(feedURL, feed)

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

func NewViewPodcastHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		pod, err := podcasts.GetPodcast(ctx, db, r.PathValue("guid"))
		if err != nil {
			// TODO handle not found error
			return err
		}

		eps, err := podcasts.ListEpisodes(ctx, db, pod.GUID)
		if err != nil {
			return err
		}

		return framework.Render(ctx, w, 200, pages.ViewPodcast(pod, eps))
	}
}

func NewDownloadPodcastHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ep, err := podcasts.GetEpisode(ctx, db, r.PathValue("guid"))
		if err != nil {
			// TODO handle not found error
			return err
		}

		f, err := os.Load(ep)
		if err != nil {
			return err
		}
		defer f.Close()

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", ep.GUID, podcasts.MimeToExt[ep.MimeType]))
		w.Header().Set("Content-Type", ep.MimeType)

		http.ServeContent(w, r, "", time.Time{}, f)
		return nil
	}
}
