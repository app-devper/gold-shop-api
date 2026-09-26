package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/gin-gonic/gin"
)

type sessionStub struct {
	userId    string
	err       error
	gotSystem string
	gotMethod string
}

func (s *sessionStub) Authorize(_ context.Context, _ string, system, method string) (string, error) {
	s.gotSystem, s.gotMethod = system, method
	return s.userId, s.err
}

func runRequireSession(stub *sessionStub, method string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", nil)
	c.Set("SessionId", "session-1")
	c.Set("System", "GOLD")
	RequireSession(stub)(c)
	return c, w
}

func TestRequireSessionSetsUserIdFromLiveSession(t *testing.T) {
	stub := &sessionStub{userId: "user-1"}
	c, w := runRequireSession(stub, http.MethodGet)
	if c.IsAborted() || c.GetString("UserId") != "user-1" {
		t.Fatalf("expected UserId user-1, got %q (status %d)", c.GetString("UserId"), w.Code)
	}
	if stub.gotSystem != "GOLD" || stub.gotMethod != http.MethodGet {
		t.Fatalf("expected token system and request method, got %q %q", stub.gotSystem, stub.gotMethod)
	}
}

func TestRequireSessionRejectsRevokedSession(t *testing.T) {
	_, w := runRequireSession(&sessionStub{err: sessionclient.ErrSessionRejected}, http.MethodGet)
	if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), "AUT-401-005") {
		t.Fatalf("expected 401 AUT-401-005, got %d %s", w.Code, w.Body.String())
	}
}

func TestRequireSessionReturns503WhenSessionStoreUnavailable(t *testing.T) {
	_, w := runRequireSession(&sessionStub{err: sessionclient.ErrUnavailable}, http.MethodPost)
	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "AUT-503-001") {
		t.Fatalf("expected 503 AUT-503-001, got %d %s", w.Code, w.Body.String())
	}
}
