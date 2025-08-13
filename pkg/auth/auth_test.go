package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/steinfletcher/apitest"
	selector "github.com/steinfletcher/apitest-css-selector"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/auth/sessions"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

func TestAuthenticationMiddleware_ValidSessionIsPassedThrough(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "Session-Id",
		Value: "validSession1", // is a fixture
	})
	resRec := &httptest.ResponseRecorder{}

	var userID uint
	var username string
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		u := users.GetUserFromCtx(ctx)
		if u != nil {
			userID = u.ID
			username = u.Username
		}
		return nil
	}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, nil)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.Empty(t, resRec.Result().Header.Get("Location"))
	assert.Equal(t, 123, int(userID))
	assert.Equal(t, "unittest", username)
}

func TestAuthenticationMiddleware_RedirectsWhenNoCookie(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	// no cookie set on request
	resRec := &httptest.ResponseRecorder{}

	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		assert.Fail(t, "next function should not be called when not authed")
		return nil
	}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, nil)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.Equal(t, "/auth/login?redirect=%2Ftest", resRec.Result().Header.Get("Location"))
}

func TestAuthenticationMiddleware_RedirectsWhenInvalidCookie(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "Session-Id",
		Value: "invalidSession1", // not in DB
	})
	resRec := &httptest.ResponseRecorder{}

	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		assert.Fail(t, "next function should not be called when not authed")
		return nil
	}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, nil)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.Equal(t, "/auth/login?redirect=%2Ftest", resRec.Result().Header.Get("Location"))
}

func TestAuthenticationMiddleware_Returns401ForHTMXRequests(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "Session-Id",
		Value: "invalidSession1", // not in DB
	})
	resRec := &httptest.ResponseRecorder{}

	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		assert.Fail(t, "next function should not be called when not authed")
		return nil
	}

	// this header is sent by HTMX
	req.Header.Add("HX-Request", "true")

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, nil)(context.Background(), resRec, req)

	assert.Equal(t, framework.HttpUnauthorized(), err)
	assert.Empty(t, resRec.Result().Header.Get("Location"))
	assert.Equal(t, "true", resRec.Result().Header.Get("HX-Refresh"))
}

func TestAuthenticationMiddleware_ValidSessionUpdatesLastSeen(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	sessionID := "validSession30MinsOld" // is a fixture

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "Session-Id",
		Value: sessionID,
	})
	resRec := &httptest.ResponseRecorder{}

	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return nil
	}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, nil)(context.Background(), resRec, req)

	assert.Nil(t, err)
	session, err := sessions.GetSessionByID(context.Background(), db, sessionID)
	if err != nil {
		panic(err)
	}
	// the fixture is well in the past, so the last seen time being recent shows it changed)
	assert.Less(t, time.Since(session.LastSeenTime), time.Minute)
	// but session created at should not change
	assert.Greater(t, time.Since(session.StartTime), time.Minute)
}

func TestAuthenticationMiddleware_SkipAuth(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	resRec := &httptest.ResponseRecorder{}

	var nextWasRun bool
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		u := users.GetUserFromCtx(ctx)
		if u != nil {
			assert.Fail(t, "unexpected user in unauthed request")
		}
		nextWasRun = true
		return nil
	}

	// tells MW to skip auth
	cfg := auth.AuthenticationMiddlewareConfig{Skip: true}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, cfg)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.Empty(t, resRec.Result().Header.Get("Location"))
	assert.True(t, nextWasRun)
}

func TestAuthenticationMiddleware_HTTPBasicAuth_ValidCredentials(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("unittest", "unittestpw") // from fixture
	resRec := &httptest.ResponseRecorder{}

	passedThrough := false
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		passedThrough = true
		return nil
	}

	// signifies HTTP basic auth should be used - i.e. for CK feeds
	cfg := auth.AuthenticationMiddlewareConfig{UseHTTPBasicAuth: true}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, cfg)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.True(t, passedThrough)
}

func TestAuthenticationMiddleware_HTTPBasicAuth_NoCredentials(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	// no credntials in request
	resRec := &httptest.ResponseRecorder{}

	passedThrough := false
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		passedThrough = true
		return nil
	}

	// signifies HTTP basic auth should be used - i.e. for CK feeds
	cfg := auth.AuthenticationMiddlewareConfig{UseHTTPBasicAuth: true}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, cfg)(context.Background(), resRec, req)

	assert.Equal(t, framework.HttpUnauthorized(), err)
	assert.False(t, passedThrough)
}

func TestAuthenticationMiddleware_HTTPBasicAuth_InvalidUsername(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("invaliduser", "password") // not in db
	resRec := &httptest.ResponseRecorder{}

	passedThrough := false
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		passedThrough = true
		return nil
	}

	// signifies HTTP basic auth should be used - i.e. for CK feeds
	cfg := auth.AuthenticationMiddlewareConfig{UseHTTPBasicAuth: true}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, cfg)(context.Background(), resRec, req)

	assert.Equal(t, framework.HttpUnauthorized(), err)
	assert.False(t, passedThrough)
}

func TestAuthenticationMiddleware_HTTPBasicAuth_InvalidPassword(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("unittest", "invalidpass") // from fixture
	resRec := &httptest.ResponseRecorder{}

	passedThrough := false
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		passedThrough = true
		return nil
	}

	// signifies HTTP basic auth should be used - i.e. for CK feeds
	cfg := auth.AuthenticationMiddlewareConfig{UseHTTPBasicAuth: true}

	mw := auth.AuthenticationMiddleware{db}
	err := mw.Handler(nextFn, cfg)(context.Background(), resRec, req)

	assert.Equal(t, framework.HttpUnauthorized(), err)
	assert.False(t, passedThrough)
}

func TestGetLoginHandler(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		err := auth.NewGetLoginHandler()(context.Background(), w, r)
		if err != nil {
			panic(err)
		}
	}

	apitest.New().
		HandlerFunc(h).
		Get("/auth/login").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Login")).
		Assert(selector.TextExists("Username")).
		Assert(selector.TextExists("Password")).
		End()
}

func TestGetLoginHandler_DisplaysLogoutMessage(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		err := auth.NewGetLoginHandler()(context.Background(), w, r)
		if err != nil {
			panic(err)
		}
	}

	apitest.New().
		HandlerFunc(h).
		Get("/auth/login").
		Query("logout", "true").
		Expect(t).
		Status(http.StatusOK).
		Assert(selector.TextExists("Login")).
		Assert(selector.TextExists("You have been logged out")).
		End()
}

func TestPostLoginHandler_ValidCredentials(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	h := func(w http.ResponseWriter, r *http.Request) {
		err := auth.NewPostLoginHandler("http://example.com", db)(context.Background(), w, r)
		if err != nil {
			panic(err)
		}
	}

	body := "username=unittest&password=unittestpw"
	length := fmt.Sprintf("%d", len(body))
	apitest.New().
		HandlerFunc(h).
		Post("/auth/login").
		Header("Content-Type", "application/x-www-form-urlencoded").
		Header("Content-Length", length).
		Body(body).
		Expect(t).
		Status(http.StatusFound).
		Header("Location", "http://example.com/").
		CookiePresent("Session-Id").
		Body("").
		End()
}

func TestPostLoginHandler_RedirectsToPath(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	h := func(w http.ResponseWriter, r *http.Request) {
		err := auth.NewPostLoginHandler("http://example.com", db)(context.Background(), w, r)
		if err != nil {
			panic(err)
		}
	}

	body := "username=unittest&password=unittestpw"
	length := fmt.Sprintf("%d", len(body))
	apitest.New().
		HandlerFunc(h).
		Post("/auth/login").
		Query("redirect", "/hello").
		Header("Content-Type", "application/x-www-form-urlencoded").
		Header("Content-Length", length).
		Body(body).
		Expect(t).
		Status(http.StatusFound).
		Header("Location", "http://example.com/hello").
		End()
}

func TestPostLoginHandler_InvalidUsername(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	h := func(w http.ResponseWriter, r *http.Request) {
		err := auth.NewPostLoginHandler("http://example.com", db)(context.Background(), w, r)
		if err != nil {
			panic(err)
		}
	}

	body := "username=invalidUser&password=password"
	length := fmt.Sprintf("%d", len(body))
	apitest.New().
		HandlerFunc(h).
		Post("/auth/login").
		Header("Content-Type", "application/x-www-form-urlencoded").
		Header("Content-Length", length).
		Body(body).
		Expect(t).
		Status(http.StatusOK).
		HeaderNotPresent("Location").
		CookieNotPresent("Session-Id").
		Assert(selector.TextExists("Login")).
		Assert(selector.TextExists("Unknown username or password")).
		End()
}

func TestPostLoginHandler_InvalidPassword(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	h := func(w http.ResponseWriter, r *http.Request) {
		err := auth.NewPostLoginHandler("http://example.com", db)(context.Background(), w, r)
		if err != nil {
			panic(err)
		}
	}

	body := "username=unittest&password=invalidPassword"
	length := fmt.Sprintf("%d", len(body))
	apitest.New().
		HandlerFunc(h).
		Post("/auth/login").
		Header("Content-Type", "application/x-www-form-urlencoded").
		Header("Content-Length", length).
		Body(body).
		Expect(t).
		Status(http.StatusOK).
		HeaderNotPresent("Location").
		CookieNotPresent("Session-Id").
		Assert(selector.TextExists("Login")).
		Assert(selector.TextExists("Unknown username or password")).
		End()
}

func TestLogoutHandler(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	h := func(w http.ResponseWriter, r *http.Request) {
		err := auth.NewLogoutHandler("http://example.com", db)(context.Background(), w, r)
		if err != nil {
			panic(err)
		}
	}

	fixtureSessionID := "validSession1"

	// verify session exists
	sBefore, err := sessions.GetSessionByID(context.Background(), db, fixtureSessionID)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, uint(123), sBefore.UserID)

	apitest.New().
		HandlerFunc(h).
		Get("/auth/logout").
		Cookie("Session-Id", fixtureSessionID).
		Expect(t).
		Status(http.StatusFound).
		Header("Location", "/auth/login?logout=true").
		Header("Set-Cookie", "Session-Id=; Path=/; Domain=example.com; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly; Secure").
		End()

	// verify session has been deleted
	_, err = sessions.GetSessionByID(context.Background(), db, fixtureSessionID)
	assert.Equal(t, "record not found", err.Error())
}
