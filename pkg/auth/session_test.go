package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/fixures"
)

func TestSessionBeforeSave(t *testing.T) {
	testCases := map[string]struct {
		session     auth.Session
		expectedErr string
	}{
		"validMin": {
			session: auth.Session{
				ID:     "session123",
				UserID: 111,
			},
			expectedErr: "",
		},
		"validFull": {
			session: auth.Session{
				ID:           "session123",
				UserID:       111,
				StartTime:    time.Now(),
				LastSeenTime: time.Now(),
			},
			expectedErr: "",
		},
		"missingID": {
			session: auth.Session{
				UserID: 111,
			},
			expectedErr: "session not valid: Key: 'Session.ID' Error:Field validation for 'ID' failed on the 'required' tag",
		},
		"idTooLong": {
			session: auth.Session{
				ID:     fixures.StrOfLen(1001),
				UserID: 111,
			},
			expectedErr: "session not valid: Key: 'Session.ID' Error:Field validation for 'ID' failed on the 'lte' tag",
		},
		"missingUserID": {
			session: auth.Session{
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
	db, resetDB := fixures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	// use session from fixture
	s, err := auth.GetSession(context.Background(), db, "validSession1")

	assert.Nil(t, err)
	assert.Equal(t, "nQSlh56p6az1uaf0fWWOSff8kcRKDUnBck0QQwfZD+I", s.ID) // sha256(validSession1)
	assert.Equal(t, 123, int(s.UserID))
	assert.Less(t, time.Since(s.StartTime), time.Minute)
	assert.Less(t, time.Since(s.LastSeenTime), time.Minute)
}

func TestGetSession_DoesNotExist(t *testing.T) {
	db, resetDB := fixures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	s, err := auth.GetSession(context.Background(), db, "notASession")

	assert.Equal(t, "record not found", err.Error())
	assert.Zero(t, s)
}

func TestGetSession_IsExpired(t *testing.T) {
	db, resetDB := fixures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	// use session from fixture
	s, err := auth.GetSession(context.Background(), db, "expiredSession1")

	assert.Equal(t, "session expired", err.Error())
	assert.Zero(t, s)
}

func TestCreateSession(t *testing.T) {
	db, resetDB := fixures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	sessionID, err := auth.CreateSession(context.Background(), db, 456)

	assert.Nil(t, err)
	assert.NotEmpty(t, sessionID)

	s, err := auth.GetSession(context.Background(), db, sessionID)

	assert.Nil(t, err)
	assert.Equal(t, 456, int(s.UserID))
}

func TestUpdateLastSeen(t *testing.T) {
	db, resetDB := fixures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	s, err := auth.GetSession(context.Background(), db, "validSession30MinsOld")
	assert.Nil(t, err)
	assert.Greater(t, time.Since(s.LastSeenTime), 15*time.Minute)

	err = auth.UpdateSessionLastSeen(context.Background(), db, &s)
	assert.Nil(t, err)

	s2, err := auth.GetSession(context.Background(), db, "validSession30MinsOld")
	assert.Nil(t, err)
	assert.Less(t, time.Since(s2.LastSeenTime), 15*time.Minute)
}

func TestDelectExpiredSessions(t *testing.T) {
	db, resetDB := fixures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	_, err := auth.GetSession(context.Background(), db, "expiredSession1")
	assert.Equal(t, "session expired", err.Error())

	n, err := auth.DeleteExpiredSessions(context.Background(), db)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), n)

	_, err = auth.GetSession(context.Background(), db, "expiredSession1")
	assert.Equal(t, "record not found", err.Error())
}
