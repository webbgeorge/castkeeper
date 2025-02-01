package webserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/web"
	"gorm.io/gorm"
)

func Start(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	feedService *podcasts.FeedService,
	db *gorm.DB,
	os objectstorage.ObjectStorage,
	itunesAPI *itunes.ItunesAPI,
) error {
	server, err := framework.NewServer(":8080", logger)
	if err != nil {
		return fmt.Errorf("failed to start server", err)
	}

	middleware := framework.DefaultMiddlewareStack()
	middleware = append(middleware, framework.NewCSRFMiddleware(
		cfg.WebServer.CSRFSecretKey,
		cfg.WebServer.CSRFSecureCookie,
	))

	return server.SetServerMiddlewares(middleware...).
		AddFileServer("GET /static/", http.FileServer(http.FS(web.StaticAssets))).
		AddRoute("GET /", NewHomeHandler(db)).
		AddRoute("GET /podcasts/{guid}", NewViewPodcastHandler(db)).
		AddRoute("GET /podcasts/search", NewSearchPodcastsHandler()).
		AddRoute("POST /partials/add-podcast", NewAddPodcastHandler(feedService, db, os)).
		AddRoute("POST /partials/search-results", NewSearchResultsHandler(itunesAPI)).
		AddRoute("GET /podcasts/{guid}/image", NewDownloadImageHandler(db, os)).
		AddRoute("GET /episodes/{guid}/download", NewDownloadEpisodeHandler(db, os)).
		AddRoute("GET /feeds/{guid}", NewFeedHandler(cfg.BaseURL, db)).
		Start(ctx)
}
