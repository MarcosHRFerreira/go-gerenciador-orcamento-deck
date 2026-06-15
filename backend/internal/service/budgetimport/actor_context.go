package budgetimport

import "context"

type actorUserIDContextKey struct{}

func WithActorUserID(ctx context.Context, userID int64) context.Context {
	if userID <= 0 {
		return ctx
	}

	return context.WithValue(ctx, actorUserIDContextKey{}, userID)
}

func actorUserIDFromContext(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}

	userID, ok := ctx.Value(actorUserIDContextKey{}).(int64)
	if !ok {
		return 0
	}

	return userID
}
