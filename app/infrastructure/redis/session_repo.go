package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// SessionRepository reads sessions stored by um-api
type SessionRepository struct {
	rdb *redis.Client
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb}
}

// GetSessionById returns the userId associated with the given sessionId
func (r *SessionRepository) GetSessionById(ctx context.Context, sessionId string) (string, error) {
	result, err := r.rdb.Get(ctx, sessionId).Result()
	if err != nil {
		return "", err
	}
	return result, nil
}
