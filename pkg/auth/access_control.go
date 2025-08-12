package auth

import (
	"context"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

type AccessControlMiddlewareConfig struct {
	RequiredAccessLevel users.AccessLevel
}

type AccessControlMiddleware struct{}

func (mw AccessControlMiddleware) Handler(next framework.Handler, config framework.MiddlewareConfig) framework.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		acConfig, ok := config.(AccessControlMiddlewareConfig)
		if !ok || acConfig.RequiredAccessLevel == 0 {
			// all endpoints must define an access level explicitly
			// if no AL is required, it should be set to -1
			framework.GetLogger(ctx).WarnContext(
				ctx, "No required access level set for route",
			)
			return framework.HttpForbidden()
		}

		if acConfig.RequiredAccessLevel == -1 {
			return next(ctx, w, r)
		}

		user := users.GetUserFromCtx(ctx)
		if user == nil {
			framework.GetLogger(ctx).WarnContext(
				ctx, "No session found for route requiring access level",
				"requiredAccessLevel", acConfig.RequiredAccessLevel,
			)
			return framework.HttpForbidden()
		}

		if !user.CheckAccessLevel(acConfig.RequiredAccessLevel) {
			framework.GetLogger(ctx).InfoContext(
				ctx, "Access control check failed",
				"userID", user.ID,
				"requiredAccessLevel", acConfig.RequiredAccessLevel,
				"accessLevel", user.AccessLevel,
			)
			return framework.HttpForbidden()
		}

		framework.GetLogger(ctx).InfoContext(
			ctx, "Access control check passed",
			"userID", user.ID,
			"requiredAccessLevel", acConfig.RequiredAccessLevel,
			"accessLevel", user.AccessLevel,
		)

		return next(ctx, w, r)
	}
}

func (mw AccessControlMiddleware) Match(config framework.MiddlewareConfig) bool {
	_, ok := config.(AccessControlMiddlewareConfig)
	return ok
}
