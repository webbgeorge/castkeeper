package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/auth/users"
	"github.com/webbgeorge/castkeeper/pkg/framework"
)

func TestAccessControlMiddleware_HigherAccessLevelIsAllowed(t *testing.T) {
	u := users.User{AccessLevel: 3}
	userCtx := users.CtxWithUser(context.Background(), u)

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(userCtx)
	resRec := &httptest.ResponseRecorder{}

	var nextCalled bool
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		nextCalled = true
		return nil
	}

	config := auth.AccessControlMiddlewareConfig{
		RequiredAccessLevel: 2,
	}

	mw := auth.AccessControlMiddleware{}
	err := mw.Handler(nextFn, config)(userCtx, resRec, req)

	assert.Nil(t, err)
	assert.True(t, nextCalled)
}

func TestAccessControlMiddleware_SameAccessLevelIsAllowed(t *testing.T) {
	u := users.User{AccessLevel: 3}
	userCtx := users.CtxWithUser(context.Background(), u)

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(userCtx)
	resRec := &httptest.ResponseRecorder{}

	var nextCalled bool
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		nextCalled = true
		return nil
	}

	config := auth.AccessControlMiddlewareConfig{
		RequiredAccessLevel: 3,
	}

	mw := auth.AccessControlMiddleware{}
	err := mw.Handler(nextFn, config)(userCtx, resRec, req)

	assert.Nil(t, err)
	assert.True(t, nextCalled)
}

func TestAccessControlMiddleware_LowerAccessLevelIsNotAllowed(t *testing.T) {
	u := users.User{AccessLevel: 2}
	userCtx := users.CtxWithUser(context.Background(), u)

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(userCtx)
	resRec := &httptest.ResponseRecorder{}

	var nextCalled bool
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		nextCalled = true
		return nil
	}

	config := auth.AccessControlMiddlewareConfig{
		RequiredAccessLevel: 3,
	}

	mw := auth.AccessControlMiddleware{}
	err := mw.Handler(nextFn, config)(userCtx, resRec, req)

	assert.Equal(t, framework.HttpForbidden(), err)
	assert.False(t, nextCalled)
}

func TestAccessControlMiddleware_ForbiddenWhenRequiredAccessLevelNotSet(t *testing.T) {
	u := users.User{AccessLevel: 2}
	userCtx := users.CtxWithUser(context.Background(), u)

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(userCtx)
	resRec := &httptest.ResponseRecorder{}

	var nextCalled bool
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		nextCalled = true
		return nil
	}

	mw := auth.AccessControlMiddleware{}
	err := mw.Handler(nextFn, nil)(userCtx, resRec, req)

	assert.Equal(t, framework.HttpForbidden(), err)
	assert.False(t, nextCalled)
}

func TestAccessControlMiddleware_ForbiddenWhenNoUserCtx(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	resRec := &httptest.ResponseRecorder{}

	var nextCalled bool
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		nextCalled = true
		return nil
	}

	config := auth.AccessControlMiddlewareConfig{
		RequiredAccessLevel: 3,
	}

	mw := auth.AccessControlMiddleware{}
	err := mw.Handler(nextFn, config)(context.Background(), resRec, req)

	assert.Equal(t, framework.HttpForbidden(), err)
	assert.False(t, nextCalled)
}

func TestAccessControlMiddleware_AllowedWhenRequiredAccessLevelSetToNegativeOne(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	resRec := &httptest.ResponseRecorder{}

	var nextCalled bool
	nextFn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		nextCalled = true
		return nil
	}

	config := auth.AccessControlMiddlewareConfig{
		RequiredAccessLevel: -1,
	}

	mw := auth.AccessControlMiddleware{}
	err := mw.Handler(nextFn, config)(context.Background(), resRec, req)

	assert.Nil(t, err)
	assert.True(t, nextCalled)
}
