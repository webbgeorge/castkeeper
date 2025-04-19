package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"gorm.io/gorm"
)

const (
	sessionIDCookie       = "Session-Id"
	authStateExpiry       = time.Hour
	sessionExpiry         = 24 * time.Hour
	sessionLastSeenExpiry = time.Hour
)

type sessionCtxKey struct{}

func GetSessionFromCtx(ctx context.Context) *Session {
	if ctx == nil {
		return nil
	}
	session, ok := ctx.Value(sessionCtxKey{}).(*Session)
	if !ok {
		return nil
	}
	return session
}

func NewAuthenticationMiddleware(db *gorm.DB) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			c, err := r.Cookie(sessionIDCookie)
			if err != nil || c.Value == "" {
				redirectToLogin(w, r)
				return nil
			}

			s, err := GetSession(ctx, db, c.Value)
			if err != nil {
				redirectToLogin(w, r)
				return nil
			}
			err = UpdateSessionLastSeen(ctx, db, &s)
			if err != nil {
				// log and continue
				framework.GetLogger(ctx).WarnContext(ctx, "failed to update session last_seen_time")
			}

			sessionCtx := context.WithValue(ctx, sessionCtxKey{}, &s)
			framework.GetLogger(ctx).InfoContext(
				ctx, "successfully authenticated user",
				"userID", s.UserID,
			)

			// TODO add user to context - should user be in session?

			return next(sessionCtx, w, r)
		}
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	queryString := url.Values{
		"redirect": []string{r.URL.Path},
	}.Encode()
	http.Redirect(w, r, fmt.Sprintf("/auth/login?%s", queryString), http.StatusFound)
}

// handles auth for podcast feeds - using basic auth and no redirecting or sessions
func NewFeedAuthenticationMiddleware(db *gorm.DB) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			username, password, ok := r.BasicAuth()
			if !ok {
				return framework.HttpUnauthorized()
			}

			user, err := checkUsernameAndPassword(ctx, db, username, password)
			if err != nil {
				return framework.HttpUnauthorized()
			}

			framework.GetLogger(ctx).InfoContext(
				ctx, "successfully authenticated user to feed",
				"userID", user.ID,
			)

			return next(ctx, w, r)
		}
	}
}

func NewGetLoginHandler() framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return renderLoginPage(ctx, w, r, false)
	}
}

func NewPostLoginHandler(
	baseURL string,
	db *gorm.DB,
) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		user, err := checkUsernameAndPassword(ctx, db, username, password)
		if err != nil {
			return renderLoginPage(ctx, w, r, true)
		}

		sID, err := CreateSession(ctx, db, user.ID)
		if err != nil {
			return err
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionIDCookie,
			Value:    sID,
			Expires:  time.Now().Add(sessionExpiry),
			Path:     "/",
			Domain:   cookieDomain(baseURL),
			Secure:   true,
			HttpOnly: true,
		})

		redirectURL := urlWithPath(baseURL, redirectPathFromReq(r))
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return nil
	}
}

func checkUsernameAndPassword(ctx context.Context, db *gorm.DB, username, password string) (User, error) {
	user, err := GetUserByUsername(ctx, db, username)
	if err != nil {
		framework.GetLogger(ctx).InfoContext(ctx, fmt.Sprintf("failed to find username: '%s'", username), "error", err)
		return User{}, err
	}

	if err := user.CheckPassword(password); err != nil {
		framework.GetLogger(ctx).InfoContext(ctx, fmt.Sprintf("incorrect password for username: '%s'", username), "error", err)
		return User{}, err
	}

	return user, nil
}

func renderLoginPage(ctx context.Context, w http.ResponseWriter, r *http.Request, authErr bool) error {
	errText := ""
	if authErr {
		errText = "Unknown username or password"
	}
	return framework.Render(ctx, w, 200, pages.Login(
		csrf.Token(r),
		redirectPathFromReq(r), // pass from GET to be rendered into a hidden input
		errText,
	))
}

func redirectPathFromReq(r *http.Request) string {
	redirect := r.PostFormValue("redirect")
	if redirect == "" {
		redirect = r.URL.Query().Get("redirect")
	}
	if redirect == "" {
		return "/"
	}
	u, err := url.Parse(redirect)
	if err != nil || u.Path == "" {
		return "/"
	}
	return sanitiseRedirectPath(u.Path)
}

func sanitiseRedirectPath(redirectPath string) string {
	if len(redirectPath) > 500 {
		return "/"
	}
	if !strings.HasPrefix(redirectPath, "/") {
		return "/"
	}
	if strings.HasPrefix(redirectPath, "/auth") {
		return "/"
	}
	return redirectPath
}

func urlWithPath(baseURL, path string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("invalid baseURL")
	}
	u.Path = path
	return u.String()
}

func cookieDomain(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic("invalid baseURL")
	}
	return u.Hostname()
}
