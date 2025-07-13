package sessions

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"gorm.io/gorm"
)

const (
	sessionIDCookie       = "Session-Id"
	sessionExpiry         = 24 * time.Hour
	sessionLastSeenExpiry = time.Hour
)

var sessionValidate = validator.New(validator.WithRequiredStructEnabled())

type Session struct {
	ID           string     `gorm:"primaryKey" validate:"required,gte=1,lte=1000"`
	UserID       uint       `validate:"required,gte=1"`
	User         users.User `validate:"-" gorm:"foreignKey:UserID"`
	StartTime    time.Time
	LastSeenTime time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (s *Session) BeforeSave(tx *gorm.DB) error {
	err := sessionValidate.Struct(s)
	if err != nil {
		return fmt.Errorf("session not valid: %w", err)
	}
	return nil
}

func GetSession(ctx context.Context, db *gorm.DB, r *http.Request) (Session, error) {
	sessionID := getSessionIDFromCookie(r)
	if sessionID == "" {
		return Session{}, errors.New("no session id provided")
	}

	return GetSessionByID(ctx, db, sessionID)
}

func GetSessionByID(ctx context.Context, db *gorm.DB, sessionID string) (Session, error) {
	var session Session
	result := db.
		Preload("User").
		First(&session, "id = ?", strSHA256(sessionID))
	if result.Error != nil {
		return session, result.Error
	}
	if time.Since(session.StartTime) > sessionExpiry || time.Since(session.LastSeenTime) > sessionLastSeenExpiry {
		return Session{}, errors.New("session expired")
	}
	return session, nil
}

func CreateSession(
	ctx context.Context,
	baseURL string,
	db *gorm.DB,
	userID uint,
	w http.ResponseWriter,
) error {
	id := uuid.New().String()
	now := time.Now()

	session := Session{
		ID:           strSHA256(id),
		UserID:       userID,
		StartTime:    now,
		LastSeenTime: now,
	}
	if err := db.Create(&session).Error; err != nil {
		return err
	}

	setSessionCookie(w, baseURL, id)

	return nil
}

func UpdateSessionLastSeen(ctx context.Context, db *gorm.DB, session *Session) error {
	now := time.Now()
	result := db.
		Model(session).
		Select("LastSeenTime").
		Updates(Session{LastSeenTime: now})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DeleteSession(
	ctx context.Context,
	baseURL string,
	db *gorm.DB,
	r *http.Request,
	w http.ResponseWriter,
) error {
	sessionID := getSessionIDFromCookie(r)
	if sessionID == "" {
		return errors.New("no session id provided")
	}

	result := db.
		Where("id = ?", strSHA256(sessionID)).
		Delete(&Session{})
	if result.Error != nil {
		return result.Error
	}

	deleteSessionCookie(w, baseURL)

	return nil
}

// DeleteExpired deletes expired sessions from the database.
// this is for cleaning unused data, and isn't part of expiring the sessions
func DeleteExpiredSessions(ctx context.Context, db *gorm.DB) (int64, error) {
	timeBuffer := time.Hour // ensure we don't delete sessions in active use
	startTimeThreshold := time.Now().Add(-sessionExpiry).Add(-timeBuffer)
	lastSeenThreshold := time.Now().Add(-sessionLastSeenExpiry).Add(-timeBuffer)

	result := db.
		Where("start_time < ? OR last_seen_time < ?", startTimeThreshold, lastSeenThreshold).
		Delete(&Session{})

	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

const HouseKeepingQueueName = "authHouseKeeping"

func NewHouseKeepingQueueWorker(
	db *gorm.DB,
) func(cxt context.Context, _ any) error {
	return func(ctx context.Context, _ any) error {
		framework.GetLogger(ctx).InfoContext(ctx, "starting auth housekeeping")

		sessionsDeleted, err := DeleteExpiredSessions(ctx, db)
		if err != nil {
			return err
		}
		framework.GetLogger(ctx).InfoContext(ctx, fmt.Sprintf("deleted '%d' sessions", sessionsDeleted))

		return nil
	}
}

func strSHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return base64.RawStdEncoding.EncodeToString(h.Sum(nil))
}

func setSessionCookie(w http.ResponseWriter, baseURL, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionIDCookie,
		Value:    sessionID,
		Expires:  time.Now().Add(sessionExpiry),
		Path:     "/",
		Domain:   cookieDomain(baseURL),
		Secure:   true,
		HttpOnly: true,
	})
}

func deleteSessionCookie(w http.ResponseWriter, baseURL string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionIDCookie,
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		Domain:   cookieDomain(baseURL),
		Secure:   true,
		HttpOnly: true,
	})
}

func getSessionIDFromCookie(r *http.Request) string {
	c, err := r.Cookie(sessionIDCookie)
	if err != nil {
		return ""
	}
	return c.Value
}

func cookieDomain(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("invalid baseURL")
	}
	return u.Hostname()
}
