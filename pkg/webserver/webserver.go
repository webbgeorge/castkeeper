package webserver

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/framework/middleware"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/web"
	"gorm.io/gorm"
)

func NewWebserver(
	cfg config.Config,
	logger *slog.Logger,
	feedService *podcasts.FeedService,
	db *gorm.DB,
	os objectstorage.ObjectStorage,
	itunesAPI *itunes.ItunesAPI,
) *framework.Server {
	port := fmt.Sprintf(":%d", cfg.WebServer.Port)
	server := framework.NewServer(port, logger)

	mw := middleware.DefaultMiddlewareStack(
		cfg.WebServer.CSRFSecretKey,
		cfg.WebServer.CSRFSecureCookie,
	)
	mw = append(
		mw,
		auth.AuthMiddleware{DB: db},
	)

	skipAuth := auth.AuthMiddlewareConfig{Skip: true}
	useBasicAuth := auth.AuthMiddlewareConfig{UseHTTPBasicAuth: true}

	return server.SetServerMiddlewares(mw...).
		AddFileServer("GET /static/", http.FileServer(http.FS(web.StaticAssets)), skipAuth).
		AddRoute("GET /", NewHomeHandler(db)).
		AddRoute("GET /auth/login", auth.NewGetLoginHandler(), skipAuth).
		AddRoute("POST /auth/login", auth.NewPostLoginHandler(cfg.BaseURL, db), skipAuth).
		AddRoute("GET /auth/logout", auth.NewLogoutHandler(cfg.BaseURL, db), skipAuth).
		AddRoute("GET /users", NewManageUsersHandler(db)).
		AddRoute("GET /users/create", NewCreateUserGetHandler(db)).
		AddRoute("POST /users/create", NewCreateUserPostHandler(db)).
		AddRoute("GET /users/{id}/edit", NewEditUserGetHandler(db)).
		AddRoute("PUT /users/{id}/username", NewUpdateUsernameHandler(db)).
		AddRoute("PUT /users/{id}/password", NewUpdatePasswordHandler(db)).
		AddRoute("POST /users/{id}/delete", NewDeleteUserHandler(db)).
		AddRoute("GET /podcasts/{guid}", NewViewPodcastHandler(cfg.BaseURL, db)).
		AddRoute("GET /podcasts/search", NewSearchPodcastsHandler()).
		AddRoute("POST /podcasts/search", NewSearchResultsHandler(itunesAPI)).
		AddRoute("POST /podcasts/add", NewAddPodcastHandler(feedService, db, os)).
		AddRoute("GET /podcasts/{guid}/image", NewDownloadImageHandler(db, os)).
		AddRoute("GET /episodes/{guid}/download", NewDownloadEpisodeHandler(db, os)).
		AddRoute("POST /episodes/{guid}/requeue-download", NewRequeueDownloadHandler(db)).
		AddRoute("GET /feeds/{guid}", NewFeedHandler(cfg.BaseURL, db), useBasicAuth)
}
