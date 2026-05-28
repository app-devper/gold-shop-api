package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mongoinfra "github.com/devper-gold/gold-shop-api/app/infrastructure/mongo"
	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

func TestRequireTenant(t *testing.T) {
	tests := []struct {
		name        string
		setClientID bool
		clientID    string
		wantStatus  int
		wantInCtx   string
	}{
		{"valid alphanumeric", true, "abc-123", http.StatusOK, "abc-123"},
		{"valid default tenant 000", true, "000", http.StatusOK, "000"},
		{"valid single char", true, "a", http.StatusOK, "a"},
		{"missing claim", false, "", http.StatusUnauthorized, ""},
		{"empty string claim", true, "", http.StatusUnauthorized, ""},
		{"path traversal", true, "../etc/passwd", http.StatusUnauthorized, ""},
		{"semicolon injection", true, "abc;drop", http.StatusUnauthorized, ""},
		{"leading underscore", true, "_abc", http.StatusUnauthorized, ""},
		{"trailing hyphen", true, "abc-", http.StatusUnauthorized, ""},
		{"space", true, "a b", http.StatusUnauthorized, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.setClientID {
				c.Set("ClientId", tc.clientID)
			}

			RequireTenant()(c)

			if w.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
			if tc.wantStatus == http.StatusOK {
				if c.IsAborted() {
					t.Error("context should not be aborted on success")
				}
				got, err := mongoinfra.ClientIDFromCtx(c.Request.Context())
				if err != nil {
					t.Fatalf("ClientIDFromCtx: %v", err)
				}
				if got != tc.wantInCtx {
					t.Errorf("ctx clientID: got %q, want %q", got, tc.wantInCtx)
				}
			} else {
				if !c.IsAborted() {
					t.Error("context should be aborted on validation failure")
				}
				if _, err := mongoinfra.ClientIDFromCtx(c.Request.Context()); err == nil {
					t.Error("request ctx should not carry clientID after rejection")
				}
			}
		})
	}
}
