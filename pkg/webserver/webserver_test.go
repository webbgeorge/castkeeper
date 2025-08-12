package webserver_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/steinfletcher/apitest"
	selector "github.com/steinfletcher/apitest-css-selector"
	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/downloadworker"
	"github.com/webbgeorge/castkeeper/pkg/feedworker"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/webserver"
	"gorm.io/gorm"
)

func TestHomePage(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
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
	ctx, server, _, _, reset := setupServerForTest()
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
	ctx, server, _, _, reset := setupServerForTest()
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
	ctx, server, _, _, reset := setupServerForTest()
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

func TestInvalidSessionOnHTMXRequestReturns401InsteadOfRedirectToLogin(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/search"). // this is a partials route
		WithContext(ctx).
		Cookie("Session-Id", "notASession").
		Header("HX-Request", "true").
		Expect(t).
		Status(http.StatusUnauthorized).
		HeaderNotPresent("Location").
		Header("HX-Refresh", "true").
		Body("").
		End()
}

func TestNotFound(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
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
	_, server, _, _, reset := setupServerForTest()
	defer reset()

	ctx := context.Background() // ctx without the csrf skip value

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/search").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("query=testPods").                // from fixtures
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusForbidden).
		End()
}

func TestSearchPodcastsPage(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
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
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/search").
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
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/search").
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
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/search").
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
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/search").
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
	ctx, server, db, root, reset := setupServerForTest()
	defer reset()

	// from fixtures, not in DB yet
	feedURL := "http://testdata/feeds/valid-not-added.xml"

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/add").
		WithContext(ctx).
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body(fmt.Sprintf("feedUrl=%s", feedURL)).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Podcast added")).
		End()

	// assert pod was added
	var podcast podcasts.Podcast
	result := db.First(&podcast, "feed_url = ?", feedURL)
	if result.Error != nil {
		panic(result.Error)
	}
	assert.Equal(t, "Test podcast 2 description goes here", podcast.Description)

	// assert image was created
	f, err := root.Open(fmt.Sprintf("%s/%s.jpg", podcast.GUID, podcast.GUID))
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	// compare against fixture content
	assert.Equal(t, "Not a real JPG", strings.TrimSpace(string(data)))

	// verify feed worker job was added to queue
	_, err = framework.PopQueueTask(ctx, db, feedworker.FeedWorkerQueueName)
	assert.Nil(t, err)
}

func TestAddPodcast_InvalidFeed(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/add").
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
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/podcasts/add").
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
	ctx, server, _, _, reset := setupServerForTest()
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
	ctx, server, _, _, reset := setupServerForTest()
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

func TestDownloadImage(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get(fmt.Sprintf("/podcasts/%s/image", genGUID("abc-123"))).
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Not a real JPG")). // fixture image has text content
		End()
}

func TestDownloadImage_NotFound(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/podcasts/not-a-pod/image").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func TestDownloadEpisode(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get(fmt.Sprintf("/episodes/%s/download", genGUID("ep-1"))).
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Not a real MP3")). // fixture mp3 has text content
		End()
}

func TestDownloadEpisode_NotFound(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/episodes/not-an-ep/download").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func TestRequeuePodcast(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post(fmt.Sprintf("/episodes/%s/requeue-download", genGUID("ep-1"))). // from fixtures
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Test episode c8998fa5-8083-56a6-8d3c-7b98d031b3d8")).
		Assert(selector.TextExists("pending")).
		End()

	// verify was added to queue
	qt, err := framework.PopQueueTask(ctx, db, downloadworker.DownloadWorkerQueueName)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "c8998fa5-8083-56a6-8d3c-7b98d031b3d8", qt.Data.(string))

	// verify that ep status was updated to pending
	ep, err := podcasts.GetEpisode(ctx, db, "c8998fa5-8083-56a6-8d3c-7b98d031b3d8")
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "pending", ep.Status)
}

func TestRequeuePodcast_NotFound(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/episodes/not-an-ep/requeue-download").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func TestGetFeed(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	expectedBody, err := os.ReadFile("./testdata/expected-generated-feed.xml")
	if err != nil {
		panic(err)
	}
	expectedBodyStr := strings.TrimSpace(string(expectedBody))

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get(fmt.Sprintf("/feeds/%s", genGUID("abc-123"))). // from fixtures
		WithContext(ctx).
		BasicAuth("unittest", "unittestpw"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Body(expectedBodyStr).
		End()
}

func TestGetFeed_NotFound(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/feeds/not-a-pod").
		WithContext(ctx).
		BasicAuth("unittest", "unittestpw"). // from fixtures
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func TestManageUserPage(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/users").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Manage users")).
		Assert(selector.ContainsTextValue(
			"table > tbody > tr:nth-child(1) > td:nth-child(1)",
			"unittest",
		)).
		End()
}

func TestCreateUserPage(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/users/create").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Create New User")).
		Assert(selector.Exists(
			"input[name=username]",
			"input[name=password]",
			"input[name=repeatPassword]",
			"select[name=accessLevel]",
		)).
		Assert(selector.ContainsTextValue("button[type=submit]", "Submit")).
		End()
}

func TestCreateUserSubmit_Success(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/create").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("username=mytestuser&password=GoodPasswordForTesting&repeatPassword=GoodPasswordForTesting&accessLevel=1").
		Expect(t).
		Status(http.StatusFound).
		Header("Location", "/users").
		End()

	// verify added to DB
	user, err := users.GetUserByUsername(ctx, db, "mytestuser")
	if err != nil {
		panic(err)
	}
	assert.NotEmpty(t, user.ID)
	assert.Nil(t, user.CheckPassword("GoodPasswordForTesting"))
}

func TestCreateUserSubmit_InvalidRequest(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/create").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("notAValid=Request").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Username is a required field")).
		End()

	// verify user is not saved
	_, err := users.GetUserByUsername(ctx, db, "mytestuser")
	assert.Equal(t, "record not found", err.Error())
}

func TestCreateUserSubmit_PasswordsDoNotMatch(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/create").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("username=mytestuser&password=GoodPasswordForTesting&repeatPassword=OtherPassword&accessLevel=1").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Passwords must match")).
		End()

	// verify user is not saved
	_, err := users.GetUserByUsername(ctx, db, "mytestuser")
	assert.Equal(t, "record not found", err.Error())
}

func TestCreateUserSubmit_PasswordTooWeak(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/create").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("username=mytestuser&password=password123&repeatPassword=password123&accessLevel=1").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("password is too easy to guess")).
		End()

	// verify user is not saved
	_, err := users.GetUserByUsername(ctx, db, "mytestuser")
	assert.Equal(t, "record not found", err.Error())
}

func TestCreateUserSubmit_DuplicateUsername(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/create").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		// username from fixtures
		Body("username=unittest&password=GoodPasswordForTesting&repeatPassword=GoodPasswordForTesting&accessLevel=1").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("A user with this username already exists")).
		End()
}

func TestEditUserPage(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Get("/users/123/edit"). //  user ID from fixtures
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Edit User")).
		Assert(selector.TextExists("Update Username")).
		Assert(selector.Exists("input[name=username][value=unittest]")).
		Assert(selector.TextExists("Update Password")).
		Assert(selector.Exists("input[name=password]", "input[name=repeatPassword]")).
		End()
}

func TestUpdateUsernameSubmit_Success(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Put("/users/456/username"). // user 'unittest2' from fixtures
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures (user 'unittest')
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("username=unittestnew").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Username was updated successfully")).
		End()

	// verify old username not in DB
	_, err := users.GetUserByUsername(ctx, db, "unittest2")
	assert.Equal(t, "record not found", err.Error())

	// verify new username is in DB
	user, err := users.GetUserByUsername(ctx, db, "unittestnew")
	if err != nil {
		panic(err)
	}
	assert.NotEmpty(t, user.ID)
}

func TestUpdateUsernameSubmit_UserNotFound(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Put("/users/999/username"). // user doesn't exist in fixtures
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures (user 'unittest')
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("username=unittestnew").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Failed to update username")).
		End()
}

func TestUpdateUsernameSubmit_InvalidUsername(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Put("/users/456/username"). // user 'unittest2' from fixtures
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures (user 'unittest')
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("username=").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Username is a required field")).
		End()
}

func TestUpdateUsernameSubmit_UsernameExists(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Put("/users/456/username"). // user 'unittest2' from fixtures
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures (user 'unittest')
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("username=unittest"). // username already exists in fixtures
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("A user with this username already exists")).
		End()
}

func TestUpdatePasswordSubmit_Success(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Put("/users/456/password").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("password=GoodPasswordForTesting&repeatPassword=GoodPasswordForTesting").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Password was updated successfully")).
		End()

	// verify updated in DB
	user, err := users.GetUserByID(ctx, db, 456)
	if err != nil {
		panic(err)
	}
	assert.NotEmpty(t, user.ID)
	assert.Nil(t, user.CheckPassword("GoodPasswordForTesting"))
}

func TestUpdatePasswordSubmit_InvalidPassword(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Put("/users/456/password").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Password is a required field, RepeatPassword is a required field")).
		End()
}

func TestUpdatePasswordSubmit_PasswordsDoNotMatch(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Put("/users/456/password").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("password=GoodPasswordForTesting&repeatPassword=NotTheSame").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Passwords must match")).
		End()
}

func TestUpdatePasswordSubmit_PasswordsTooWeak(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Put("/users/456/password").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("password=password123&repeatPassword=password123").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("password is too easy to guess")).
		End()
}

func TestDeleteUser_Success(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/456/delete").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("").
		Expect(t).
		Status(http.StatusOK).
		HeaderNotPresent("HX-Trigger").
		End()

	// verify deleted in DB
	_, err := users.GetUserByID(ctx, db, 456)
	assert.Equal(t, "record not found", err.Error())
}

func TestDeleteUser_InvalidUserID(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/invalidIDFormat/delete").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("").
		Expect(t).
		Status(http.StatusOK).
		Header("HX-Trigger", `{"showMessage":"Invalid user ID in request"}`).
		End()
}

func TestDeleteUser_CannotDeleteCurrentUser(t *testing.T) {
	ctx, server, db, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/123/delete"). // 123 is current user in fixtures
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("").
		Expect(t).
		Status(http.StatusOK).
		Header("HX-Trigger", `{"showMessage":"Cannot delete the current user"}`).
		End()

	// verify not deleted in DB
	_, err := users.GetUserByID(ctx, db, 123)
	assert.Nil(t, err)
}

func TestDeleteUser_UserNotFound(t *testing.T) {
	ctx, server, _, _, reset := setupServerForTest()
	defer reset()

	apitest.New().
		HandlerFunc(server.Mux.ServeHTTP).
		Post("/users/999/delete").
		WithContext(ctx).
		Cookie("Session-Id", "validSession1"). // from fixtures
		Header("Content-Type", "application/x-www-form-urlencoded").
		Body("").
		Expect(t).
		Status(http.StatusNotFound).
		End()
}

func setupServerForTest() (context.Context, *framework.Server, *gorm.DB, *os.Root, func()) {
	db := fixtures.ConfigureDBForTestWithFixtures()
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
	root, resetFS := fixtures.ConfigureFSForTestWithFixtures()
	os := &objectstorage.LocalObjectStorage{
		Root:       root,
		HTTPClient: fixtures.TestDataHTTPClient,
	}
	itunesAPI := &itunes.ItunesAPI{
		HTTPClient: fixtures.TestItunesHTTPClient,
	}

	server := webserver.NewWebserver(cfg, logger, feedService, db, os, itunesAPI)
	ctx := context.WithValue(context.Background(), "gorilla.csrf.Skip", true)

	return ctx, server, db, root, func() {
		resetFS()
	}
}

func genGUID(s string) string {
	return uuid.NewV5(uuid.NamespaceOID, s).String()
}
