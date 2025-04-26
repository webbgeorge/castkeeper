package webserver_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gofrs/uuid/v5"
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
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("CastKeeper")).
		Assert(selector.TextExists("Your Podcasts")).
		Assert(selector.TextExists("Test podcast 916ed63b-7e5e-5541-af78-e214a0c14d95")). // from fixtures
		End()
}

func TestNoSessionRedirectsToLogin(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/").
		WithContext(ctx).
		Expect(t).
		Status(http.StatusFound).
		Header("Location", "/auth/login?redirect=%2F").
		Body("").
		End()
}

func TestExpiredSessionRedirectsToLogin(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/podcasts/search").
		WithContext(ctx).
		Cookie("Session-Id", "expiredSession1"). // from fixtures
		Expect(t).
		Status(http.StatusFound).
		Header("Location", "/auth/login?redirect=%2Fpodcasts%2Fsearch").
		Body("").
		End()
}

func TestInvalidSessionRedirectsToLogin(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/").
		WithContext(ctx).
		Cookie("Session-Id", "notASession").
		Expect(t).
		Status(http.StatusFound).
		Header("Location", "/auth/login?redirect=%2F").
		Body("").
		End()
}

func TestNotFound(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/notAPage").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusNotFound).
		Assert(selector.TextExists("Not Found")).
		End()
}

func TestCSRFFailure(t *testing.T) {
	server, _, reset := setupServerForTest()
	defer reset()

	ctx := context.Background() // ctx without the csrf skip value

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/partials/search-results").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("query=testPods").                // from fixtures
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusForbidden).
		End()
}

func TestSearchPodcastsPage(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/podcasts/search").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Add Podcast")).
		Assert(selector.Exists(`input[type="text"][name="query"]`)).
		Assert(selector.ContainsTextValue("button", "Search")).
		Assert(selector.ContainsTextValue("button", "Add Feed URL")).
		End()
}

func TestSearchResults_Success(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/partials/search-results").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("query=testPods").                // from fixtures
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Elis James and John Robins")).
		Assert(selector.TextExists("Elis James and John Robins on Radio X Podcast")).
		End()
}

func TestSearchResults_EmptyResults(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/partials/search-results").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("query=noTestPods").              // from fixtures
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("You may want to try different keywords or checking for typos.")).
		End()
}

func TestSearchResults_InvalidQuery(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/partials/search-results").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("query=").
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Search query cannot be empty")).
		End()
}

func TestSearchResults_FailedToCallItunes(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/partials/search-results").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("query=500").                     // from fixtures
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("There was an unexpected error")).
		End()
}

func TestAddPodcast_Success(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/partials/add-podcast").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("feedUrl=http://testdata/feeds/very-long-pod-title.xml"). // from fixtures, not in DB yet
		Cookie("Session-Id", "validSession1").                         // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Podcast added")).
		End()

	// TODO assert that image was saved
}

func TestAddPodcast_InvalidFeed(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/partials/add-podcast").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("feedUrl=http://testdata/feeds/invalid.xml"). // from fixtures
		Cookie("Session-Id", "validSession1").             // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Invalid feed")).
		End()
}

func TestAddPodcast_AlreadyAdded(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/partials/add-podcast").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("feedUrl=http://testdata/feeds/valid.xml"). // from fixtures, already in db
		Cookie("Session-Id", "validSession1").           // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("This podcast is already added")).
		End()
}

func TestViewPodcast(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get(fmt.Sprintf("/podcasts/%s", genGUID("abc-123"))). // from fixtures
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Test podcast 916ed63b-7e5e-5541-af78-e214a0c14d95")).
		Assert(selector.TextExists("Dr Tester")).
		Assert(selector.TextExists("2 episodes")).
		Assert(selector.TextExists("Test podcast description goes here")).
		Assert(selector.TextExists("Test episode c8998fa5-8083-56a6-8d3c-7b98d031b3d8")).
		Assert(selector.TextExists("Test episode 3864ebe7-7a8f-5532-841f-0bacd0a0cc6c")).
		End()
}

func TestViewPodcast_NotFound(t *testing.T) {
	server, ctx, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/podcasts/not-a-pod").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusNotFound).
		Assert(selector.TextExists("404 Not Found")).
		End()
}

// TODO download podcast
// TODO requeue podcast
// TODO download image
// TODO feed

func setupServerForTest() (*framework.Server, context.Context, func()) {
	db, resetDB := fixtures.ConfigureDBForTestWithFixtures()
	cfg := config.Config{
		BaseURL: "http://example.com",
		WebServer: config.WebServerConfig{
			Port:             8000,
			CSRFSecretKey:    "testValueDoNotUseInProd",
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
	ctx := context.WithValue(context.Background(), "gorilla.csrf.Skip", true)

	return server, ctx, resetDB
}

func genGUID(s string) string {
	return uuid.NewV5(uuid.NamespaceOID, s).String()
}
