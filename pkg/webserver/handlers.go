package webserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/components/partials"
	"github.com/webbgeorge/castkeeper/pkg/downloadworker"
	"github.com/webbgeorge/castkeeper/pkg/feedworker"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/util"
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
		return framework.Render(ctx, w, 200, pages.SearchPodcasts(csrf.Token(r)))
	}
}

func NewSearchResultsHandler(itunesAPI *itunes.ItunesAPI) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		q := r.PostFormValue("query")
		if len(q) == 0 {
			return framework.Render(ctx, w, 200, partials.SearchResults(csrf.Token(r), nil, "Search query cannot be empty"))
		}
		if len(q) >= 250 {
			return framework.Render(ctx, w, 200, partials.SearchResults(csrf.Token(r), nil, "Search query must be less than 250 characters"))
		}

		results, err := itunesAPI.Search(ctx, q)
		if err != nil {
			framework.GetLogger(ctx).ErrorContext(ctx, "itunes search failed", "error", err)
			return framework.Render(ctx, w, 200, partials.SearchResults(csrf.Token(r), nil, "There was an unexpected error"))
		}
		return framework.Render(ctx, w, 200, partials.SearchResults(csrf.Token(r), results, ""))
	}
}

func NewAddPodcastHandler(feedService *podcasts.FeedService, db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		feedURL := r.PostFormValue("feedUrl")
		podcast, _, err := feedService.ParseFeed(ctx, feedURL)
		if err != nil {
			if !errors.Is(err, podcasts.ParseErrors{}) {
				framework.GetLogger(ctx).ErrorContext(ctx, "error parsing feed", "error", err)
				return framework.Render(ctx, w, 200, partials.AddPodcast("Invalid feed"))
			}
			framework.GetLogger(ctx).WarnContext(ctx, fmt.Sprintf("some episodes of podcast '%s' had parsing errors: %s", podcast.GUID, err.Error()))
			// continue even with some episode parse failures...
		}

		if err = db.Create(&podcast).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return framework.Render(ctx, w, 200, partials.AddPodcast("This podcast is already added"))
			}
			return err
		}

		// TODO detect filetype
		fileName := fmt.Sprintf("%s.%s", util.SanitiseGUID(podcast.GUID), "jpg")
		_, err = os.SaveRemoteFile(ctx, podcast.ImageURL, util.SanitiseGUID(podcast.GUID), fileName)
		if err != nil {
			framework.GetLogger(ctx).WarnContext(ctx, "failed to download image, continuing without", "error", err)
		}

		err = framework.PushQueueTask(ctx, db, feedworker.FeedWorkerQueueName, "")
		if err != nil {
			framework.GetLogger(ctx).WarnContext(ctx, "failed to queue feed worker, continuing without", "error", err)
		}

		return framework.Render(ctx, w, 200, partials.AddPodcast(""))
	}
}

func NewViewPodcastHandler(baseURL string, db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		pod, err := podcasts.GetPodcast(ctx, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		eps, err := podcasts.ListEpisodes(ctx, db, pod.GUID)
		if err != nil {
			return err
		}

		return framework.Render(ctx, w, 200, pages.ViewPodcast(csrf.Token(r), baseURL, pod, eps))
	}
}

func NewDownloadEpisodeHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ep, err := podcasts.GetEpisode(ctx, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		extension, err := podcasts.MIMETypeExtension(ep.MimeType)
		if err != nil {
			return err
		}

		w.Header().Set(
			"Content-Disposition",
			fmt.Sprintf("attachment; filename=%s.%s", ep.GUID, extension),
		)
		w.Header().Set("Content-Type", ep.MimeType)

		fileName := fmt.Sprintf("%s.%s", util.SanitiseGUID(ep.GUID), extension)
		return os.ServeFile(ctx, r, w, util.SanitiseGUID(ep.PodcastGUID), fileName)
	}
}

func NewRequeueDownloadHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ep, err := podcasts.GetEpisode(ctx, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			err := podcasts.UpdateEpisodeStatus(ctx, tx, &ep, podcasts.EpisodeStatusPending, nil)
			if err != nil {
				return err
			}
			err = framework.PushQueueTask(ctx, tx, downloadworker.DownloadWorkerQueueName, ep.GUID)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		ep.Status = podcasts.EpisodeStatusPending
		return framework.Render(ctx, w, 200, partials.EpisodeListItem(csrf.Token(r), ep))
	}
}

func NewDownloadImageHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		pod, err := podcasts.GetPodcast(ctx, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		// TODO
		// w.Header().Set("Content-Type", ep.MimeType)

		// TODO detect file type
		fileName := fmt.Sprintf("%s.%s", util.SanitiseGUID(pod.GUID), "jpg")
		return os.ServeFile(ctx, r, w, util.SanitiseGUID(pod.GUID), fileName)
	}
}

func NewFeedHandler(baseURL string, db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		feed, err := podcasts.GenerateFeed(ctx, baseURL, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		w.Header().Set("Content-Type", "application/xml")
		return feed.WriteFeedXML(w)
	}
}
