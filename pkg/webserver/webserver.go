package webserver

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
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
		auth.AuthenticationMiddleware{DB: db},
		auth.AccessControlMiddleware{},
	)

	skipAuth := auth.AuthenticationMiddlewareConfig{Skip: true}
	useBasicAuth := auth.AuthenticationMiddlewareConfig{UseHTTPBasicAuth: true}

	requireNone := auth.AccessControlMiddlewareConfig{RequiredAccessLevel: users.AccessLevelNone}
	requireReadOnly := auth.AccessControlMiddlewareConfig{RequiredAccessLevel: users.AccessLevelReadOnly}
	requireManagePods := auth.AccessControlMiddlewareConfig{RequiredAccessLevel: users.AccessLevelManagePodcasts}
	requireAdmin := auth.AccessControlMiddlewareConfig{RequiredAccessLevel: users.AccessLevelAdmin}

	return server.SetServerMiddlewares(mw...).
		AddFileServer("GET /static/", http.FileServer(http.FS(web.StaticAssets)), skipAuth, requireNone).
		AddRoute("GET /", NewHomeHandler(db), requireReadOnly).
		AddRoute("GET /auth/login", auth.NewGetLoginHandler(), skipAuth, requireNone).
		AddRoute("POST /auth/login", auth.NewPostLoginHandler(cfg.BaseURL, db), skipAuth, requireNone).
		AddRoute("GET /auth/logout", auth.NewLogoutHandler(cfg.BaseURL, db), skipAuth, requireNone).
		AddRoute("GET /profile/password", NewCurrentUserUpdatePasswordGetHandler(db), requireReadOnly).
		AddRoute("POST /profile/password", NewCurrentUserUpdatePasswordPostHandler(db), requireReadOnly).
		AddRoute("GET /users", NewManageUsersHandler(db), requireAdmin).
		AddRoute("GET /users/create", NewCreateUserGetHandler(db), requireAdmin).
		AddRoute("POST /users/create", NewCreateUserPostHandler(db), requireAdmin).
		AddRoute("GET /users/{id}/edit", NewEditUserGetHandler(db), requireAdmin).
		AddRoute("PUT /users/{id}", NewUpdateUserHandler(db), requireAdmin).
		AddRoute("PUT /users/{id}/password", NewUpdatePasswordHandler(db), requireAdmin).
		AddRoute("POST /users/{id}/delete", NewDeleteUserHandler(db), requireAdmin).
		AddRoute("GET /podcasts/{guid}", NewViewPodcastHandler(cfg.BaseURL, db), requireReadOnly).
		AddRoute("GET /podcasts/search", NewSearchPodcastsHandler(), requireManagePods).
		AddRoute("POST /podcasts/search", NewSearchResultsHandler(itunesAPI), requireManagePods).
		AddRoute("POST /podcasts/add", NewAddPodcastHandler(feedService, db, os), requireManagePods).
		AddRoute("GET /podcasts/{guid}/image", NewDownloadImageHandler(db, os), requireReadOnly).
		AddRoute("GET /episodes/{guid}/download", NewDownloadEpisodeHandler(db, os), requireReadOnly).
		AddRoute("POST /episodes/{guid}/requeue-download", NewRequeueDownloadHandler(db), requireManagePods).
		AddRoute("GET /feeds/{guid}", NewFeedHandler(cfg.BaseURL, db), useBasicAuth, requireReadOnly)
}
