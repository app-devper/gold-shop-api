package redis

import (
	"context"
	"errors"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/sirupsen/logrus"
)

// SessionRepository confirms the UM session behind a verified access token by
// reading UM's session store through sessionclient (um-api ADR-0004).
type SessionRepository struct {
	checker *sessionclient.Checker
}

// NewSessionRepository reads UM sessions from the Redis at hostOrURL. If that
// is missing or invalid it fails closed: every check reports the store as
// unavailable.
func NewSessionRepository(hostOrURL string) *SessionRepository {
	checker, err := sessionclient.New(hostOrURL)
	if err != nil || !checker.Enabled() {
		logrus.Errorf("session client not configured (REDIS_HOST): %v; every session check will fail", err)
		return &SessionRepository{}
	}
	return &SessionRepository{checker: checker}
}

// Authorize returns the session's user id. It fails with
// sessionclient.ErrSessionRejected when the session is gone or belongs to
// another system, and with sessionclient.ErrUnavailable when the store cannot
// answer and the outage policy does not let the request continue.
func (r *SessionRepository) Authorize(ctx context.Context, sessionId, system, method string) (string, error) {
	if r.checker == nil {
		return "", sessionclient.ErrUnavailable
	}
	session, err := r.checker.Authorize(ctx, sessionId, system, method)
	if err != nil {
		if !errors.Is(err, sessionclient.ErrSessionRejected) {
			logrus.Warn("session check: ", err)
		}
		return "", err
	}
	return session.UserId, nil
}
