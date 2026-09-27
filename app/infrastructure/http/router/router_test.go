package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/app-devper/um-api/sessionclient"
	"github.com/devper-gold/gold-shop-api/app/infrastructure/http/middleware"
	"github.com/gin-gonic/gin"
)

// unusedStore stands in for UM's session store; anonymous requests never
// reach it.
type unusedStore struct{}

func (unusedStore) Session(context.Context, string) (sessionclient.Session, error) {
	return sessionclient.Session{}, sessionclient.ErrUnavailable
}

// Every /api/gold/v1 route must refuse a request without a UM token before a
// handler or repository runs, so the nil handlers and repositories here are
// never used.
func TestEveryProtectedRouteRejectsAnAnonymousRequest(t *testing.T) {
	auth, err := middleware.NewAuthWithStore("test-secret", "GOLD", unusedStore{})
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Setup(r, auth, "", nil, nil, &Handlers{})

	checked := 0
	for _, route := range r.Routes() {
		if !strings.HasPrefix(route.Path, "/api/gold/v1") {
			continue
		}
		checked++
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(route.Method, strings.ReplaceAll(route.Path, ":", "x"), nil))
		if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), sessionclient.CodeMissingToken) {
			t.Errorf("%s %s: expected 401 %s without a token, got %d %s",
				route.Method, route.Path, sessionclient.CodeMissingToken, w.Code, w.Body.String())
		}
	}
	if checked == 0 {
		t.Fatal("no /api/gold/v1 routes were checked")
	}
	t.Logf("checked %d protected routes", checked)
}
