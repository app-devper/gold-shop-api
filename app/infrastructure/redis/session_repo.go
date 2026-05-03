package redis

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

// um-api stores sessions under "session:<id>" with a JSON body keyed by userId.
const sessionPrefix = "session:"

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
	raw, err := r.rdb.Get(ctx, sessionPrefix+sessionId).Result()
	if err != nil {
		return "", err
	}
	var data struct {
		UserId string `json:"userId"`
	}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return "", err
	}
	return data.UserId, nil
}
