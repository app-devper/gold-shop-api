package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func signToken(t *testing.T, system string) string {
	t.Helper()
	claims := AccessClaims{
		Role:     "ADMIN",
		System:   system,
		ClientId: "000",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "session-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func runRequireAuthenticated(authorization string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if authorization != "" {
		c.Request.Header.Set("Authorization", authorization)
	}
	RequireAuthenticated(testSecret, "GOLD")(c)
	return c, w
}

func TestRequireAuthenticatedAcceptsTokenForThisSystem(t *testing.T) {
	c, w := runRequireAuthenticated("Bearer " + signToken(t, "GOLD"))
	if c.IsAborted() {
		t.Fatalf("expected token to pass, got %d %s", w.Code, w.Body.String())
	}
	if c.GetString("SessionId") != "session-1" || c.GetString("System") != "GOLD" || c.GetString("ClientId") != "000" {
		t.Fatalf("expected claims in context, got session=%q system=%q client=%q",
			c.GetString("SessionId"), c.GetString("System"), c.GetString("ClientId"))
	}
}

func TestRequireAuthenticatedRejectsTokenForAnotherSystem(t *testing.T) {
	c, w := runRequireAuthenticated("Bearer " + signToken(t, "POS"))
	if !c.IsAborted() || w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), "AUT-401-006") {
		t.Fatalf("expected 401 AUT-401-006 for a POS token, got %d %s", w.Code, w.Body.String())
	}
	if c.GetString("SessionId") != "" {
		t.Fatal("rejected token must not set a session in context")
	}
}

func TestRequireAuthenticatedRejectsMissingOrBadToken(t *testing.T) {
	for name, header := range map[string]string{
		"missing":    "",
		"not bearer": "Basic abc",
		"wrong secret": "Bearer " + func() string {
			s, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, AccessClaims{System: "GOLD",
				RegisteredClaims: jwt.RegisteredClaims{ID: "x"}}).SignedString([]byte("other"))
			return s
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			c, w := runRequireAuthenticated(header)
			if !c.IsAborted() || w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d %s", w.Code, w.Body.String())
			}
		})
	}
}
