package webserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/web"
	"go.opentelemetry.io/otel/sdk/resource"
	"gorm.io/gorm"
)

func Start(
	ctx context.Context,
	otelRes *resource.Resource,
	db *gorm.DB,
	os objectstorage.ObjectStorage,
	itunesAPI *itunes.ItunesAPI,
) error {
	server, err := framework.NewServer(otelRes, ":8080")
	if err != nil {
		return fmt.Errorf("failed to start server", err)
	}

	middleware := framework.DefaultMiddlewareStack()

	return server.SetServerMiddlewares(middleware...).
		AddFileServer("GET /static/", http.FileServer(http.FS(web.StaticAssets))).
		AddRoute("GET /", NewHomeHandler(db)).
		AddRoute("GET /podcasts/subscribe", NewSubscribeHandler()).
		AddRoute("POST /partials/podcasts/subscribe", NewSubmitSubscribeHandler(db, os)).
		AddRoute("GET /podcasts/{guid}", NewViewPodcastHandler(db)).
		AddRoute("GET /podcasts/{guid}/image", NewDownloadImageHandler(db, os)).
		AddRoute("GET /episodes/{guid}/download", NewDownloadPodcastHandler(db, os)).
		AddRoute("POST /partials/search", NewSearchPostHandler(itunesAPI)).
		Start(ctx)
}
