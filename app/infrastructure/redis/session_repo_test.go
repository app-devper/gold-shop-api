package redis

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/app-devper/um-api/sessionclient"
)

// Without REDIS_HOST the repository must fail closed, never authorize.
func TestUnconfiguredSessionRepositoryFailsClosed(t *testing.T) {
	repo := NewSessionRepository("")
	userId, err := repo.Authorize(context.Background(), "s1", "GOLD", http.MethodGet)
	if !errors.Is(err, sessionclient.ErrUnavailable) || userId != "" {
		t.Fatalf("expected ErrUnavailable and no user, got %q %v", userId, err)
	}
}
