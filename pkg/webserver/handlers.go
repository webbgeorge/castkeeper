package webserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gorilla/csrf"
	"github.com/gorilla/schema"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/components/partials"
	"github.com/webbgeorge/castkeeper/pkg/downloadworker"
	"github.com/webbgeorge/castkeeper/pkg/feedworker"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/itunes"
	"github.com/webbgeorge/castkeeper/pkg/objectstorage"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/pkg/util"
	"gorm.io/gorm"
)

var (
	decoder    = schema.NewDecoder()
	validate   = validator.New(validator.WithRequiredStructEnabled())
	uni        = ut.New(en.New(), en.New())
	enTrans, _ = uni.GetTranslator("en")
)

func init() {
	decoder.IgnoreUnknownKeys(true)
	_ = en_translations.RegisterDefaultTranslations(validate, enTrans)
}

func NewHomeHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		if r.URL.Path != "/" {
			// handle fallback on home route
			return framework.HttpNotFound()
		}

		pods, err := podcasts.ListPodcasts(ctx, db)
		if err != nil {
			return err
		}
		return framework.Render(ctx, w, 200, pages.Home(pods))
	}
}

func NewSearchPodcastsHandler() framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return framework.Render(ctx, w, 200, pages.SearchPodcasts(csrf.Token(r)))
	}
}

func NewSearchResultsHandler(itunesAPI *itunes.ItunesAPI) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		q := r.PostFormValue("query")
		if len(q) == 0 {
			return framework.Render(ctx, w, 200, partials.SearchResults(csrf.Token(r), nil, "Search query cannot be empty"))
		}
		if len(q) >= 250 {
			return framework.Render(ctx, w, 200, partials.SearchResults(csrf.Token(r), nil, "Search query must be less than 250 characters"))
		}

		results, err := itunesAPI.Search(ctx, q)
		if err != nil {
			framework.GetLogger(ctx).ErrorContext(ctx, "itunes search failed", "error", err)
			return framework.Render(ctx, w, 200, partials.SearchResults(csrf.Token(r), nil, "There was an unexpected error"))
		}
		return framework.Render(ctx, w, 200, partials.SearchResults(csrf.Token(r), results, ""))
	}
}

func NewAddPodcastHandler(feedService *podcasts.FeedService, db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		feedURL := r.PostFormValue("feedUrl")
		podcast, _, err := feedService.ParseFeed(ctx, feedURL)
		if err != nil {
			if !errors.Is(err, podcasts.ParseErrors{}) {
				framework.GetLogger(ctx).ErrorContext(ctx, "error parsing feed", "error", err)
				return framework.Render(ctx, w, 200, partials.AddPodcast("Invalid feed"))
			}
			framework.GetLogger(ctx).WarnContext(ctx, fmt.Sprintf("some episodes of podcast '%s' had parsing errors: %s", podcast.GUID, err.Error()))
			// continue even with some episode parse failures...
		}

		if err = db.Create(&podcast).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return framework.Render(ctx, w, 200, partials.AddPodcast("This podcast is already added"))
			}
			return err
		}

		// TODO detect filetype
		fileName := fmt.Sprintf("%s.%s", util.SanitiseGUID(podcast.GUID), "jpg")
		_, err = os.SaveRemoteFile(ctx, podcast.ImageURL, util.SanitiseGUID(podcast.GUID), fileName)
		if err != nil {
			framework.GetLogger(ctx).WarnContext(ctx, "failed to download image, continuing without", "error", err)
		}

		err = framework.PushQueueTask(ctx, db, feedworker.FeedWorkerQueueName, "")
		if err != nil {
			framework.GetLogger(ctx).WarnContext(ctx, "failed to queue feed worker, continuing without", "error", err)
		}

		return framework.Render(ctx, w, 200, partials.AddPodcast(""))
	}
}

func NewViewPodcastHandler(baseURL string, db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		pod, err := podcasts.GetPodcast(ctx, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		eps, err := podcasts.ListEpisodes(ctx, db, pod.GUID)
		if err != nil {
			return err
		}

		return framework.Render(ctx, w, 200, pages.ViewPodcast(csrf.Token(r), baseURL, pod, eps))
	}
}

func NewDownloadEpisodeHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ep, err := podcasts.GetEpisode(ctx, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		extension, err := podcasts.MIMETypeExtension(ep.MimeType)
		if err != nil {
			return err
		}

		w.Header().Set(
			"Content-Disposition",
			fmt.Sprintf("attachment; filename=%s.%s", ep.GUID, extension),
		)
		w.Header().Set("Content-Type", ep.MimeType)

		fileName := fmt.Sprintf("%s.%s", util.SanitiseGUID(ep.GUID), extension)
		return os.ServeFile(ctx, r, w, util.SanitiseGUID(ep.PodcastGUID), fileName)
	}
}

func NewRequeueDownloadHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ep, err := podcasts.GetEpisode(ctx, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			err := podcasts.UpdateEpisodeStatus(ctx, tx, &ep, podcasts.EpisodeStatusPending, nil)
			if err != nil {
				return err
			}
			err = framework.PushQueueTask(ctx, tx, downloadworker.DownloadWorkerQueueName, ep.GUID)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		ep.Status = podcasts.EpisodeStatusPending
		return framework.Render(ctx, w, 200, partials.EpisodeListItem(csrf.Token(r), ep))
	}
}

func NewDownloadImageHandler(db *gorm.DB, os objectstorage.ObjectStorage) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		pod, err := podcasts.GetPodcast(ctx, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		// TODO
		// w.Header().Set("Content-Type", ep.MimeType)

		// TODO detect file type
		fileName := fmt.Sprintf("%s.%s", util.SanitiseGUID(pod.GUID), "jpg")
		return os.ServeFile(ctx, r, w, util.SanitiseGUID(pod.GUID), fileName)
	}
}

func NewFeedHandler(baseURL string, db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		feed, err := podcasts.GenerateFeed(ctx, baseURL, db, r.PathValue("guid"))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		w.Header().Set("Content-Type", "application/xml")
		return feed.WriteFeedXML(w)
	}
}

func NewManageUsersHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		users, err := users.ListUsers(ctx, db)
		if err != nil {
			return err
		}
		createUserSuccess := r.URL.Query().Get("createUserSuccess") == "true"
		return framework.Render(ctx, w, 200, pages.ManageUsers(pages.ManageUsersViewModel{
			CSRFToken:         csrf.Token(r),
			Users:             users,
			CreateUserSuccess: createUserSuccess,
		}))
	}
}

func NewDeleteUserHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		currentUser := users.GetUserFromCtx(ctx)

		userID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil || userID == 0 {
			framework.GetLogger(ctx).Info("cannot delete user: invalid user ID in request")
			setShowMessageHeader(w, "Invalid user ID in request", "error")
			w.Header().Add("HX-Reswap", "none")
			w.WriteHeader(http.StatusOK)
			return nil
		}

		if uint(userID) == currentUser.ID {
			framework.GetLogger(ctx).Info("cannot delete the current user")
			setShowMessageHeader(w, "Cannot delete the current user", "error")
			w.Header().Add("HX-Reswap", "none")
			w.WriteHeader(http.StatusOK)
			return nil
		}

		if err := users.DeleteUser(ctx, db, uint(userID)); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return framework.HttpNotFound()
			}
			return err
		}

		setShowMessageHeader(w, "User deleted successfully", "success")
		w.WriteHeader(http.StatusOK)

		return nil
	}
}

func NewCreateUserGetHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		return framework.Render(ctx, w, 200, pages.CreateUser(pages.CreateUserViewModel{
			CSRFToken: csrf.Token(r),
		}))
	}
}

func NewCreateUserPostHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		renderPage := func(formData pages.CreateUserFormData, errorText string) error {
			return framework.Render(ctx, w, 200, pages.CreateUser(
				pages.CreateUserViewModel{
					CSRFToken: csrf.Token(r),
					ErrorText: errorText,
					FormData:  formData,
				},
			))
		}

		var formData pages.CreateUserFormData
		err := parseFormData(r, &formData)
		if err != nil {
			return renderPage(formData, "Invalid request")
		}

		err = validate.Struct(formData)
		if err != nil {
			if errorText, ok := translateValidationErrs(err); ok {
				return renderPage(formData, errorText)
			}
			return renderPage(formData, "Invalid request")
		}

		if formData.Password != formData.RepeatPassword {
			return renderPage(formData, "Passwords must match")
		}

		err = users.CreateUser(
			ctx,
			db,
			formData.Username,
			formData.Password,
			formData.AccessLevel,
		)
		if err != nil {
			if pErr, ok := err.(users.PasswordStrengthError); ok {
				return renderPage(formData, pErr.Error())
			}
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return renderPage(formData, "A user with this username already exists")
			}
			framework.GetLogger(ctx).Error(
				fmt.Sprintf("failed to create user: %s", err.Error()),
			)
			return renderPage(formData, "Failed to create user")
		}

		http.Redirect(w, r, "/users?createUserSuccess=true", http.StatusFound)
		return nil
	}
}

func NewEditUserGetHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		userID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			return framework.HttpBadRequest("Invalid request URL")
		}

		user, err := users.GetUserByID(ctx, db, uint(userID))
		if err != nil {
			return fmt.Errorf("failed to GetUserByID: %w", err)
		}

		return framework.Render(ctx, w, 200, pages.EditUser(pages.EditUserViewModel{
			UpdateUserFormVM: partials.UpdateUserFormViewModel{
				CSRFToken: csrf.Token(r),
				UserID:    uint(userID),
				FormData: partials.UpdateUserFormData{
					Username:    user.Username,
					AccessLevel: user.AccessLevel,
				},
			},
			UpdatePasswordFormVM: partials.UpdatePasswordFormViewModel{
				CSRFToken: csrf.Token(r),
				UserID:    uint(userID),
			},
		}))
	}
}

func NewUpdateUserHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		userID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			return framework.HttpBadRequest("Invalid request URL")
		}

		renderPage := func(formData partials.UpdateUserFormData, errorText string, isSuccess bool) error {
			return framework.Render(ctx, w, 200, partials.UpdateUserForm(
				partials.UpdateUserFormViewModel{
					CSRFToken: csrf.Token(r),
					ErrorText: errorText,
					IsSuccess: isSuccess,
					UserID:    uint(userID),
					FormData:  formData,
				},
			))
		}

		var formData partials.UpdateUserFormData
		err = parseFormData(r, &formData)
		if err != nil {
			return renderPage(formData, "Invalid request", false)
		}

		err = validate.Struct(formData)
		if err != nil {
			if errorText, ok := translateValidationErrs(err); ok {
				return renderPage(formData, errorText, false)
			}
			return renderPage(formData, "Invalid request", false)
		}

		if uint(userID) == users.GetUserFromCtx(ctx).ID &&
			formData.AccessLevel < users.AccessLevelAdmin {
			return renderPage(formData, "To prevent lockout, you cannot change the current user access level to below Admin", false)
		}

		err = users.UpdateUser(ctx, db, uint(userID), formData.Username, formData.AccessLevel)
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return renderPage(formData, "A user with this username already exists", false)
			}
			framework.GetLogger(ctx).Error(fmt.Sprintf(
				"failed to update user '%d': %s",
				userID,
				err.Error(),
			))
			return renderPage(formData, "Failed to update user", false)
		}

		return renderPage(partials.UpdateUserFormData{
			Username:    formData.Username,
			AccessLevel: formData.AccessLevel,
		}, "", true)
	}
}

func NewUpdatePasswordHandler(db *gorm.DB) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		userID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			return framework.HttpBadRequest("Invalid request URL")
		}

		renderPage := func(formData partials.UpdatePasswordFormData, errorText string, isSuccess bool) error {
			return framework.Render(ctx, w, 200, partials.UpdatePasswordForm(
				partials.UpdatePasswordFormViewModel{
					CSRFToken: csrf.Token(r),
					ErrorText: errorText,
					IsSuccess: isSuccess,
					UserID:    uint(userID),
					FormData:  formData,
				},
			))
		}

		var formData partials.UpdatePasswordFormData
		err = parseFormData(r, &formData)
		if err != nil {
			return renderPage(formData, "Invalid request", false)
		}

		err = validate.Struct(formData)
		if err != nil {
			if errorText, ok := translateValidationErrs(err); ok {
				return renderPage(formData, errorText, false)
			}
			return renderPage(formData, "Invalid request", false)
		}

		if formData.Password != formData.RepeatPassword {
			return renderPage(formData, "Passwords must match", false)
		}

		err = users.UpdatePassword(ctx, db, uint(userID), formData.Password)
		if err != nil {
			if pErr, ok := err.(users.PasswordStrengthError); ok {
				return renderPage(formData, pErr.Error(), false)
			}
			framework.GetLogger(ctx).Error(fmt.Sprintf(
				"failed to update password for user '%d': %s",
				userID,
				err.Error(),
			))
			return renderPage(formData, "Failed to update password", false)
		}

		return renderPage(partials.UpdatePasswordFormData{}, "", true)
	}
}

func parseFormData(r *http.Request, formData any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = decoder.Decode(formData, r.PostForm)
	if err != nil {
		return err
	}

	return nil
}

// TODO move to validation package
func translateValidationErrs(err error) (string, bool) {
	errorTexts := make([]string, 0)

	var vErr validator.ValidationErrors
	if !errors.As(err, &vErr) {
		return "", false
	}
	for _, e := range vErr {
		errorTexts = append(errorTexts, e.Translate(enTrans))
	}
	if len(errorTexts) == 0 {
		return "", false
	}
	return strings.Join(errorTexts, ", "), true
}

func setShowMessageHeader(w http.ResponseWriter, message, level string) {
	w.Header().Set(
		"HX-Trigger",
		fmt.Sprintf(
			`{"showMessage":{"level":"%s","message":"%s"}}`,
			level,
			message,
		),
	)
}
