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
		s := auth.GetSessionFromCtx(ctx)
		if s != nil {
			userID = s.UserID
			username = s.User.Username
		}
		return nil
	}

	mw := auth.NewAuthenticationMiddleware(db)
	err := mw(nextFn)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.Empty(t, resRec.Header().Get("Location"))
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

	mw := auth.NewAuthenticationMiddleware(db)
	err := mw(nextFn)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.Equal(t, "/auth/login?redirect=%2Ftest", resRec.Header().Get("Location"))
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

	mw := auth.NewAuthenticationMiddleware(db)
	err := mw(nextFn)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.Equal(t, "/auth/login?redirect=%2Ftest", resRec.Header().Get("Location"))
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

	mw := auth.NewAuthenticationMiddleware(db)
	err := mw(nextFn)(context.Background(), resRec, req)

	assert.Nil(t, err)
	session, err := auth.GetSession(context.Background(), db, sessionID)
	if err != nil {
		panic(err)
	}
	// the fixture is well in the past, so the last seen time being recent shows it changed)
	assert.Less(t, time.Since(session.LastSeenTime), time.Minute)
	// but session created at should not change
	assert.Greater(t, time.Since(session.StartTime), time.Minute)
}

func TestFeedAuthenticationMiddleware_ValidCredentials(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("unittest", "unittestpw") // from fixture
	resRec := &httptest.ResponseRecorder{}

	passedThrough := false
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		passedThrough = true
		return nil
	}

	mw := auth.NewFeedAuthenticationMiddleware(db)
	err := mw(nextFn)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.True(t, passedThrough)
}

func TestFeedAuthenticationMiddleware_NoCredentials(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	// no credntials in request
	resRec := &httptest.ResponseRecorder{}

	passedThrough := false
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		passedThrough = true
		return nil
	}

	mw := auth.NewFeedAuthenticationMiddleware(db)
	err := mw(nextFn)(context.Background(), resRec, req)

	assert.Equal(t, framework.HttpUnauthorized(), err)
	assert.False(t, passedThrough)
}

func TestFeedAuthenticationMiddleware_InvalidUsername(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("invaliduser", "password") // not in db
	resRec := &httptest.ResponseRecorder{}

	passedThrough := false
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		passedThrough = true
		return nil
	}

	mw := auth.NewFeedAuthenticationMiddleware(db)
	err := mw(nextFn)(context.Background(), resRec, req)

	assert.Equal(t, framework.HttpUnauthorized(), err)
	assert.False(t, passedThrough)
}

func TestFeedAuthenticationMiddleware_InvalidPassword(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("unittest", "invalidpass") // from fixture
	resRec := &httptest.ResponseRecorder{}

	passedThrough := false
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		passedThrough = true
		return nil
	}

	mw := auth.NewFeedAuthenticationMiddleware(db)
	err := mw(nextFn)(context.Background(), resRec, req)

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
