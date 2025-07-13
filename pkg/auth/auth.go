package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/webbgeorge/castkeeper/pkg/auth/sessions"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"gorm.io/gorm"
)

func NewAuthenticationMiddleware(db *gorm.DB, redirectToLoginOnUnauth bool) framework.Middleware {
	return func(next framework.Handler) framework.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			s, err := sessions.GetSession(ctx, db, r)
			if err != nil {
				return unauthResponse(w, r, redirectToLoginOnUnauth)
			}
			err = sessions.UpdateSessionLastSeen(ctx, db, &s)
			if err != nil {
				// log and continue
				framework.GetLogger(ctx).WarnContext(ctx, "failed to update session last_seen_time")
			}

			sessionCtx := sessions.CtxWithSession(ctx, s)
			framework.GetLogger(ctx).InfoContext(
				ctx, "successfully authenticated user",
				"userID", s.UserID,
			)

			return next(sessionCtx, w, r)
		}
	}
}

func unauthResponse(w http.ResponseWriter, r *http.Request, redirectToLoginOnUnauth bool) error {
	if redirectToLoginOnUnauth {
		redirectToLogin(w, r)
		return nil
	}
	return framework.HttpUnauthorized()
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

		err = sessions.CreateSession(ctx, baseURL, db, user.ID, w)
		if err != nil {
			return err
		}

		redirectURL := urlWithPath(baseURL, redirectPathFromReq(r))
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return nil
	}
}

func NewLogoutHandler(
	baseURL string,
	db *gorm.DB,
) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		err := sessions.DeleteSession(ctx, baseURL, db, r, w)
		if err != nil {
			return err
		}
		http.Redirect(w, r, "/auth/login?logout=true", http.StatusFound)
		return nil
	}
}

func checkUsernameAndPassword(ctx context.Context, db *gorm.DB, username, password string) (users.User, error) {
	user, err := users.GetUserByUsername(ctx, db, username)
	if err != nil {
		framework.GetLogger(ctx).InfoContext(ctx, fmt.Sprintf("failed to find username: '%s'", username), "error", err)
		return users.User{}, err
	}

	if err := user.CheckPassword(password); err != nil {
		framework.GetLogger(ctx).InfoContext(ctx, fmt.Sprintf("incorrect password for username: '%s'", username), "error", err)
		return users.User{}, err
	}

	return user, nil
}

func renderLoginPage(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	authErr bool,
) error {
	errText := ""
	if authErr {
		errText = "Unknown username or password"
	}

	// query param is set when redirecting to login page from logout
	isLogout := r.URL.Query().Get("logout") == "true"

	return framework.Render(ctx, w, 200, pages.Login(
		csrf.Token(r),
		redirectPathFromReq(r), // pass from GET to be rendered into a hidden input
		errText,
		isLogout,
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
