package sessions

import "context"

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

func CtxWithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey{}, &s)
}
