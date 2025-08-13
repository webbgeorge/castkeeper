package users

import "context"

type userCtxKey struct{}

func GetUserFromCtx(ctx context.Context) *User {
	if ctx == nil {
		return nil
	}
	user, ok := ctx.Value(userCtxKey{}).(*User)
	if !ok {
		return nil
	}
	return user
}

func CtxWithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, &u)
}
