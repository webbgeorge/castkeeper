package webserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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

func NewSearchPodcastsHandler() framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return framework.Render(ctx, w, 200, pages.SearchPodcasts())
	}
}

func NewAddPodcastHandler(feedService *podcasts.FeedService, db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		feedURL := r.PostFormValue("feedUrl")
		feed, err := feedService.ParseFeed(ctx, feedURL)
		if err != nil {
			framework.GetLogger(ctx).ErrorContext(ctx, "error parsing feed", "error", err)
			return framework.Render(ctx, w, 200, partials.AddPodcast("Invalid feed"))
		}

		podcast := podcasts.PodcastFromFeed(feedURL, feed)

		if err = db.Create(&podcast).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return framework.Render(ctx, w, 200, partials.AddPodcast("This podcast is already added"))
			}
			return err
		}

		// TODO detect filetype
		fileName := fmt.Sprintf("%s.%s", podcast.GUID, "jpg")
		err = os.SaveRemoteFile(ctx, podcast.ImageURL, podcast.GUID, fileName)
		if err != nil {
			framework.GetLogger(ctx).WarnContext(ctx, "failed to download image, continuing without", "error", err)
		}

		return framework.Render(ctx, w, 200, partials.AddPodcast(""))
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

func NewDownloadEpisodeHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ep, err := podcasts.GetEpisode(ctx, db, r.PathValue("guid"))
		if err != nil {
			// TODO handle not found error
			return err
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", ep.GUID, podcasts.MimeToExt[ep.MimeType]))
		w.Header().Set("Content-Type", ep.MimeType)

		fileName := fmt.Sprintf("%s.%s", ep.GUID, podcasts.MimeToExt[ep.MimeType])
		return os.ServeFile(ctx, r, w, ep.PodcastGUID, fileName)
	}
}

func NewDownloadImageHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		pod, err := podcasts.GetPodcast(ctx, db, r.PathValue("guid"))
		if err != nil {
			// TODO handle not found error
			return err
		}

		// TODO
		// w.Header().Set("Content-Type", ep.MimeType)

		// TODO detect file type
		fileName := fmt.Sprintf("%s.%s", pod.GUID, "jpg")
		return os.ServeFile(ctx, r, w, pod.GUID, fileName)
	}
}

func NewSearchResultsHandler(itunesAPI *itunes.ItunesAPI) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		q := r.PostFormValue("query")
		if len(q) == 0 {
			return framework.Render(ctx, w, 200, partials.SearchResults(nil, "Search query cannot be empty"))
		}
		if len(q) >= 250 {
			return framework.Render(ctx, w, 200, partials.SearchResults(nil, "Search query must be less than 250 characters"))
		}

		results, err := itunesAPI.Search(ctx, q)
		if err != nil {
			framework.GetLogger(ctx).ErrorContext(ctx, "itunes search failed", "error", err)
			return framework.Render(ctx, w, 200, partials.SearchResults(nil, "There was an unexpected error"))
		}
		return framework.Render(ctx, w, 200, partials.SearchResults(results, ""))
	}
}
