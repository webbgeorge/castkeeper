package fixtures

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log/slog"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/auth/sessions"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func ConfigureDBForTestWithFixtures() *gorm.DB {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db, err := database.ConfigureDatabase(config.Config{}, logger, true)
	if err != nil {
		panic(err)
	}

	applyFixtures(db)

	return db
}

func applyFixtures(db *gorm.DB) {
	// pod with eps success
	pod, eps := podFixture("http://testdata/feeds/valid.xml")
	create(db, &pod)
	for _, ep := range eps {
		ep.Status = podcasts.EpisodeStatusSuccess
		create(db, &ep)
	}

	// pod with eps pending
	pod, eps = podFixture("http://testdata/feeds/valid-eps-pending.xml")
	create(db, &pod)
	for _, ep := range eps {
		ep.Status = podcasts.EpisodeStatusPending
		create(db, &ep)
	}

	create(db, userFixture(123, "unittest", "unittestpw", users.AccessLevelAdmin))
	create(db, userFixture(456, "unittest2", "unittestpw2", users.AccessLevelAdmin))
	create(db, userFixture(789, "readonly1", "unittestpw3", users.AccessLevelReadOnly))
	create(db, sessionFixture(
		"validSession1",
		123,
		time.Now(),
		time.Now(),
	))
	thirtyMinsAgo := time.Now().Add(-1 * time.Minute * 30)
	create(db, sessionFixture(
		"validSession30MinsOld",
		123,
		thirtyMinsAgo,
		thirtyMinsAgo,
	))
	aTimeInThePast, err := time.Parse(time.RFC3339, "2024-12-25T12:00:00Z")
	if err != nil {
		panic(err)
	}
	create(db, sessionFixture(
		"expiredSession1",
		123,
		aTimeInThePast,
		aTimeInThePast,
	))
	create(db, sessionFixture(
		"validSessionReadOnly",
		789,
		time.Now(),
		time.Now(),
	))
}

func create(db *gorm.DB, value any) {
	if err := db.Create(value).Error; err != nil {
		panic(err)
	}
}

func podFixture(feedURL string) (podcasts.Podcast, []podcasts.Episode) {
	// use the testdata feed service to get fixture podcast for convenience
	feedService := podcasts.FeedService{
		HTTPClient: TestDataHTTPClient,
	}
	pod, eps, err := feedService.ParseFeed(context.Background(), feedURL)
	if err != nil {
		panic(err)
	}
	return pod, eps
}

func userFixture(id uint, username, password string, accessLevel users.AccessLevel) *users.User {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}
	return &users.User{
		Model:       gorm.Model{ID: id},
		Username:    username,
		Password:    string(passwordHash),
		AccessLevel: accessLevel,
	}
}

func sessionFixture(id string, userID uint, startTime, seenTime time.Time) *sessions.Session {
	h := sha256.New()
	h.Write([]byte(id))
	idHash := base64.RawStdEncoding.EncodeToString(h.Sum(nil))
	return &sessions.Session{
		ID:           idHash,
		UserID:       userID,
		StartTime:    startTime,
		LastSeenTime: seenTime,
	}
}

func randomHex() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
