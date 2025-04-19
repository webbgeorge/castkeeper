package fixures

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/database"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func ConfigureDBForTestWithFixtures() (db *gorm.DB, resetFn func()) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	dbPath := path.Join(fixtureDir(), fmt.Sprintf("../../data/test-unit-%s.db", randomHex()))
	db, err := database.ConfigureDatabase(config.Config{
		Database: config.DatabaseConfig{
			Driver: "sqlite",
			DSN:    dbPath,
		},
	}, logger)
	if err != nil {
		panic(err)
	}

	applyFixtures(db)

	return db, func() {
		_ = os.Remove(dbPath)
	}
}

func applyFixtures(db *gorm.DB) {
	pod, eps := podFixture()
	create(db, &pod)
	for _, ep := range eps {
		ep.Status = podcasts.EpisodeStatusSuccess
		create(db, &ep)
	}

	create(db, userFixture(123, "unittest", "unittestpw"))
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
}

func create(db *gorm.DB, value any) {
	if err := db.Create(value).Error; err != nil {
		panic(err)
	}
}

func podFixture() (podcasts.Podcast, []podcasts.Episode) {
	// use the testdata feed service to get fixture podcast for convenience
	feedService := podcasts.FeedService{
		HTTPClient: TestDataHTTPClient,
	}
	pod, eps, err := feedService.ParseFeed(context.Background(), "http://testdata/feeds/valid.xml")
	if err != nil {
		panic(err)
	}
	return pod, eps
}

func userFixture(id uint, username, password string) *auth.User {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}
	return &auth.User{
		Model:    gorm.Model{ID: id},
		Username: username,
		Password: string(passwordHash),
	}
}

func sessionFixture(id string, userID uint, startTime, seenTime time.Time) *auth.Session {
	h := sha256.New()
	h.Write([]byte(id))
	idHash := base64.RawStdEncoding.EncodeToString(h.Sum(nil))
	return &auth.Session{
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
