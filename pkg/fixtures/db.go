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
	"github.com/webbgeorge/castkeeper/pkg/database/encryption"
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
	evs := ConfigureEncryptedValueServiceForTest()

	applyFixtures(db, evs)

	return db
}

func applyFixtures(db *gorm.DB, evs *encryption.EncryptedValueService) {
	// pod with eps success
	podFixture(db, evs, "http://testdata/feeds/valid.xml", nil, podcasts.EpisodeStatusSuccess)

	// pod with eps pending
	podFixture(db, evs, "http://testdata/feeds/valid-eps-pending.xml", nil, podcasts.EpisodeStatusPending)

	// authenticated pod
	podFixture(db, evs, "http://testdata/authenticated/feeds/valid.xml", &podcasts.PodcastCredentials{
		Username: AuthenticatedFeedCreds.Username,
		Password: AuthenticatedFeedCreds.Password,
	}, podcasts.EpisodeStatusPending)

	userFixture(db, 123, "unittest", "unittestpw", users.AccessLevelAdmin)
	userFixture(db, 456, "unittest2", "unittestpw2", users.AccessLevelAdmin)
	userFixture(db, 789, "readonly1", "unittestpw3", users.AccessLevelReadOnly)

	sessionFixture(db, "validSession1", 123, time.Now(), time.Now())
	thirtyMinsAgo := time.Now().Add(-1 * time.Minute * 30)
	sessionFixture(db, "validSession30MinsOld", 123, thirtyMinsAgo, thirtyMinsAgo)
	aTimeInThePast, err := time.Parse(time.RFC3339, "2024-12-25T12:00:00Z")
	if err != nil {
		panic(err)
	}
	sessionFixture(db, "expiredSession1", 123, aTimeInThePast, aTimeInThePast)
	sessionFixture(db, "validSessionReadOnly", 789, time.Now(), time.Now())
}

func create(db *gorm.DB, value any) {
	if err := db.Create(value).Error; err != nil {
		panic(err)
	}
}

func podFixture(
	db *gorm.DB,
	evs *encryption.EncryptedValueService,
	feedURL string,
	creds *podcasts.PodcastCredentials,
	epStatus string,
) {
	feedService := &podcasts.FeedService{
		HTTPClient: TestDataHTTPClient,
	}
	_, err := podcasts.AddPodcast(
		context.Background(), db, feedService, evs, feedURL, creds)
	if err != nil {
		panic(err)
	}
	_, eps, err := feedService.ParseFeed(context.Background(), feedURL, creds)
	if err != nil {
		panic(err)
	}
	for _, ep := range eps {
		ep.Status = epStatus
		create(db, &ep)
	}
}

func userFixture(db *gorm.DB, id uint, username, password string, accessLevel users.AccessLevel) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}
	create(db, &users.User{
		Model:       gorm.Model{ID: id},
		Username:    username,
		Password:    string(passwordHash),
		AccessLevel: accessLevel,
	})
}

func sessionFixture(db *gorm.DB, id string, userID uint, startTime, seenTime time.Time) {
	h := sha256.New()
	h.Write([]byte(id))
	idHash := base64.RawStdEncoding.EncodeToString(h.Sum(nil))
	create(db, &sessions.Session{
		ID:           idHash,
		UserID:       userID,
		StartTime:    startTime,
		LastSeenTime: seenTime,
	})
}

func randomHex() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
