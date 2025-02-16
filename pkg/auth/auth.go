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
			if err != nil || c.Value == "" {
				redirectToLogin(w, r)
				return nil
			}
			err = UpdateSessionLastSeen(ctx, db, &s)
			if err != nil {
				// log and continue
				framework.GetLogger(ctx).WarnContext(ctx, "failed to update session last_seen_time")
			}

			sessionCtx := context.WithValue(ctx, sessionCtxKey{}, s)
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

func NewGetLoginHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return framework.Render(ctx, w, 200, pages.Login(csrf.Token(r)))
	}
}

func NewPostLoginHandler(
	baseURL string,
	db *gorm.DB,
) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		user, err := GetUserByUsername(ctx, db, username)
		if err != nil {
			framework.GetLogger(ctx).InfoContext(ctx, fmt.Sprintf("failed to find username: '%s'", username), "error", err)
			// TODO log error message and show in UI
			return framework.Render(ctx, w, 200, pages.Login(csrf.Token(r)))
		}

		if err := user.checkPassword(password); err != nil {
			framework.GetLogger(ctx).InfoContext(ctx, fmt.Sprintf("incorrect password for username: '%s'", username), "error", err)
			// TODO log error message and show in UI
			return framework.Render(ctx, w, 200, pages.Login(csrf.Token(r)))
		}

		sID, err := CreateSession(ctx, db, user.ID)
		if err != nil {
			framework.GetLogger(ctx).ErrorContext(ctx, "failed to create session", "error", err)
			// TODO log error message and show in UI
			return framework.Render(ctx, w, 200, pages.Login(csrf.Token(r)))
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

func redirectPathFromReq(r *http.Request) string {
	if r.URL.Query().Get("redirect") == "" {
		return "/"
	}
	u, err := url.Parse(r.URL.Query().Get("redirect"))
	if err != nil || u.Path == "" {
		return "/"
	}
	return sanitiseRedirectPath(u.Path)
}

func sanitiseRedirectPath(redirectPath string) string {
	if len(redirectPath) > 100 {
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
