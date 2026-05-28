package mongo

import (
	"context"
	"errors"
)

type ctxKey int

const clientIDCtxKey ctxKey = iota

var ErrMissingClientID = errors.New("clientId missing from context")

func WithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDCtxKey, clientID)
}

func ClientIDFromCtx(ctx context.Context) (string, error) {
	v, ok := ctx.Value(clientIDCtxKey).(string)
	if !ok || v == "" {
		return "", ErrMissingClientID
	}
	return v, nil
}
