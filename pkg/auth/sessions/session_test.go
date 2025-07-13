package sessions_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/auth/sessions"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
)

func TestSessionBeforeSave(t *testing.T) {
	testCases := map[string]struct {
		session     sessions.Session
		expectedErr string
	}{
		"validMin": {
			session: sessions.Session{
				ID:     "session123",
				UserID: 111,
			},
			expectedErr: "",
		},
		"validFull": {
			session: sessions.Session{
				ID:           "session123",
				UserID:       111,
				StartTime:    time.Now(),
				LastSeenTime: time.Now(),
			},
			expectedErr: "",
		},
		"missingID": {
			session: sessions.Session{
				UserID: 111,
			},
			expectedErr: "session not valid: Key: 'Session.ID' Error:Field validation for 'ID' failed on the 'required' tag",
		},
		"idTooLong": {
			session: sessions.Session{
				ID:     fixtures.StrOfLen(1001),
				UserID: 111,
			},
			expectedErr: "session not valid: Key: 'Session.ID' Error:Field validation for 'ID' failed on the 'lte' tag",
		},
		"missingUserID": {
			session: sessions.Session{
				ID: "session123",
			},
			expectedErr: "session not valid: Key: 'Session.UserID' Error:Field validation for 'UserID' failed on the 'required' tag",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.session.BeforeSave(nil)
			if tc.expectedErr == "" {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.expectedErr, err.Error())
			}
		})
	}
}

func TestGetSession_IsValid(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	r := httptest.NewRequest("GET", "/", nil)
	// use session from fixture
	r.AddCookie(&http.Cookie{
		Name:  "Session-Id",
		Value: "validSession1",
	})

	s, err := sessions.GetSession(context.Background(), db, r)

	assert.Nil(t, err)
	assert.Equal(t, "nQSlh56p6az1uaf0fWWOSff8kcRKDUnBck0QQwfZD+I", s.ID) // sha256(validSession1)
	assert.Equal(t, 123, int(s.UserID))
	assert.Less(t, time.Since(s.StartTime), time.Minute)
	assert.Less(t, time.Since(s.LastSeenTime), time.Minute)
}

func TestGetSession_NoCookie(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	// request with no session cookie
	r := httptest.NewRequest("GET", "/", nil)

	_, err := sessions.GetSession(context.Background(), db, r)

	assert.Equal(t, "no session id provided", err.Error())
}

func TestGetSession_EmptyCookie(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  "Session-Id",
		Value: "",
	})

	_, err := sessions.GetSession(context.Background(), db, r)

	assert.Equal(t, "no session id provided", err.Error())
}

func TestGetSession_InvalidID(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{
		Name:  "Session-Id",
		Value: "invalidSessionID",
	})

	_, err := sessions.GetSession(context.Background(), db, r)

	assert.Equal(t, "record not found", err.Error())
}

func TestGetSessionByID_IsValid(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	// use session from fixture
	s, err := sessions.GetSessionByID(context.Background(), db, "validSession1")

	assert.Nil(t, err)
	assert.Equal(t, "nQSlh56p6az1uaf0fWWOSff8kcRKDUnBck0QQwfZD+I", s.ID) // sha256(validSession1)
	assert.Equal(t, 123, int(s.UserID))
	assert.Less(t, time.Since(s.StartTime), time.Minute)
	assert.Less(t, time.Since(s.LastSeenTime), time.Minute)
}

func TestGetSessionBtID_DoesNotExist(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	s, err := sessions.GetSessionByID(context.Background(), db, "notASession")

	assert.Equal(t, "record not found", err.Error())
	assert.Zero(t, s)
}

func TestGetSessionByID_IsExpired(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	// use session from fixture
	s, err := sessions.GetSessionByID(context.Background(), db, "expiredSession1")

	assert.Equal(t, "session expired", err.Error())
	assert.Zero(t, s)
}

func TestCreateSession(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	w := httptest.NewRecorder()

	err := sessions.CreateSession(
		context.Background(),
		"https://example.com",
		db,
		456,
		w,
	)
	assert.Nil(t, err)

	sessionIDCookie, err := http.ParseSetCookie(w.Header().Get("Set-Cookie"))
	if err != nil {
		panic(err)
	}

	assert.NotEmpty(t, sessionIDCookie.Value)
	assert.Equal(t, "/", sessionIDCookie.Path)
	assert.Equal(t, "example.com", sessionIDCookie.Domain)
	assert.True(t, sessionIDCookie.Secure)
	assert.True(t, sessionIDCookie.HttpOnly)
	assert.True(t, sessionIDCookie.Expires.After(time.Now()))

	s, err := sessions.GetSessionByID(context.Background(), db, sessionIDCookie.Value)

	assert.Nil(t, err)
	assert.Equal(t, 456, int(s.UserID))
}

func TestUpdateLastSeen(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	s, err := sessions.GetSessionByID(context.Background(), db, "validSession30MinsOld")
	assert.Nil(t, err)
	assert.Greater(t, time.Since(s.LastSeenTime), 15*time.Minute)

	err = sessions.UpdateSessionLastSeen(context.Background(), db, &s)
	assert.Nil(t, err)

	s2, err := sessions.GetSessionByID(context.Background(), db, "validSession30MinsOld")
	assert.Nil(t, err)
	assert.Less(t, time.Since(s2.LastSeenTime), 15*time.Minute)
}

func TestDeleteSession(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	fixtureSessionID := "validSession1"

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	// use session from fixture
	r.AddCookie(&http.Cookie{
		Name:  "Session-Id",
		Value: fixtureSessionID,
	})

	// assert session exists at the start
	_, err := sessions.GetSessionByID(context.Background(), db, fixtureSessionID)
	assert.Nil(t, err)

	err = sessions.DeleteSession(context.Background(), "https://example.com", db, r, w)
	assert.Nil(t, err)

	// assert session does not exists at the end
	_, err = sessions.GetSessionByID(context.Background(), db, fixtureSessionID)
	assert.Equal(t, "record not found", err.Error())

	// assert cookie was removed
	assert.Equal(
		t,
		"Session-Id=; Path=/; Domain=example.com; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly; Secure",
		w.Header().Get("Set-Cookie"),
	)
}

func TestDeleteExpiredSessions(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	_, err := sessions.GetSessionByID(context.Background(), db, "expiredSession1")
	assert.Equal(t, "session expired", err.Error())

	n, err := sessions.DeleteExpiredSessions(context.Background(), db)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), n)

	_, err = sessions.GetSessionByID(context.Background(), db, "expiredSession1")
	assert.Equal(t, "record not found", err.Error())
}

func TestHouseKeepingQueueWorker(t *testing.T) {
	db := fixtures.ConfigureDBForTestWithFixtures()

	_, err := sessions.GetSessionByID(context.Background(), db, "expiredSession1")
	assert.Equal(t, "session expired", err.Error())

	err = sessions.NewHouseKeepingQueueWorker(db)(context.Background(), "")
	assert.Nil(t, err)

	_, err = sessions.GetSessionByID(context.Background(), db, "expiredSession1")
	assert.Equal(t, "record not found", err.Error())
}
