package webserver_test

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/steinfletcher/apitest"
	selector "github.com/steinfletcher/apitest-css-selector"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/webserver"
)

func TestHomePage(t *testing.T) {
	server, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/").
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("CastKeeper")).
		Assert(selector.TextExists("Your Podcasts")).
		Assert(selector.TextExists("Test podcast 916ed63b-7e5e-5541-af78-e214a0c14d95")). // from fixtures
		End()
}

func setupServerForTest() (*framework.Server, func()) {
	db, resetDB := fixtures.ConfigureDBForTestWithFixtures()
	cfg := config.Config{
		BaseURL: "http://example.com",
		WebServer: config.WebServerConfig{
			Port:             8000,
			CSRFSecretKey:    "testKey",
			CSRFSecureCookie: false,
		},
	}
	logger := slog.New(slog.DiscardHandler)
	feedService := &podcasts.FeedService{
		HTTPClient: fixtures.TestDataHTTPClient,
	}
	os := &objectstorage.LocalObjectStorage{} // TODO
	itunesAPI := &itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	server := webserver.NewWebserver(cfg, logger, feedService, db, os, itunesAPI)

	return server, resetDB
}
