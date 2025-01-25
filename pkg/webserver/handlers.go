package webserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/components/partials"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
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

func NewSubscribeHandler() framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return framework.Render(ctx, w, 200, pages.Subscribe())
	}
}

func NewSubmitSubscribeHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		feedURL := r.PostFormValue("feedUrl")
		feed, err := podcasts.ParseFeed(ctx, feedURL)
		if err != nil {
			// TODO log err
			return framework.Render(ctx, w, 200, partials.SubscribeSubmit("Invalid feed"))
		}

		podcast := podcasts.PodcastFromFeed(feedURL, feed)

		if err = db.Create(&podcast).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return framework.Render(ctx, w, 200, partials.SubscribeSubmit("Already subscribed to this feed"))
			}
			return err
		}

		err = os.DownloadImageFromSource(podcast)
		if err != nil {
			// TODO log warning and continue
		}

		return framework.Render(ctx, w, 200, partials.SubscribeSubmit(""))
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

		f, err := os.LoadEpisode(ep)
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

func NewDownloadImageHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		pod, err := podcasts.GetPodcast(ctx, db, r.PathValue("guid"))
		if err != nil {
			// TODO handle not found error
			return err
		}

		f, err := os.LoadImage(pod)
		if err != nil {
			return err
		}
		defer f.Close()

		// TODO
		// w.Header().Set("Content-Type", ep.MimeType)

		http.ServeContent(w, r, "", time.Time{}, f)
		return nil
	}
}

func NewSearchPostHandler(itunesAPI *itunes.ItunesAPI) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		q := r.PostFormValue("query")
		if len(q) == 0 {
			return framework.Render(ctx, w, 200, partials.Search(nil, "Search query cannot be empty"))
		}
		if len(q) >= 250 {
			return framework.Render(ctx, w, 200, partials.Search(nil, "Search query must be less than 250 characters"))
		}

		results, err := itunesAPI.Search(q)
		if err != nil {
			// TODO log err
			return framework.Render(ctx, w, 200, partials.Search(nil, "There was an unexpected error"))
		}
		return framework.Render(ctx, w, 200, partials.Search(results, ""))
	}
}
