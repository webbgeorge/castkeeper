package webserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/web"
	"gorm.io/gorm"
)

func Start(
	ctx context.Context,
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

	return server.SetServerMiddlewares(middleware...).
		AddFileServer("GET /static/", http.FileServer(http.FS(web.StaticAssets))).
		AddRoute("GET /", NewHomeHandler(db)).
		AddRoute("GET /podcasts/subscribe", NewSubscribeHandler()).
		AddRoute("POST /partials/podcasts/subscribe", NewSubmitSubscribeHandler(feedService, db, os)).
		AddRoute("GET /podcasts/{guid}", NewViewPodcastHandler(db)).
		AddRoute("GET /podcasts/{guid}/image", NewDownloadImageHandler(db, os)).
		AddRoute("GET /episodes/{guid}/download", NewDownloadPodcastHandler(db, os)).
		AddRoute("POST /partials/search", NewSearchPostHandler(itunesAPI)).
		Start(ctx)
}
