package mongo

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestValidateClientID(t *testing.T) {
	maxLen := strings.Repeat("a", 50)
	tooLong := strings.Repeat("a", 51)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"default tenant", "000", false},
		{"alpha only", "abc", false},
		{"single char", "a", false},
		{"mixed alphanumeric", "Shop1", false},
		{"with hyphen", "shop-01", false},
		{"with underscore", "a_b_c", false},
		{"max length 50", maxLen, false},

		{"empty", "", true},
		{"path traversal", "../etc", true},
		{"semicolon injection", "abc;drop", true},
		{"space", "a b", true},
		{"dot", "a.b", true},
		{"slash", "a/b", true},
		{"dollar", "a$b", true},
		{"unicode", "ทดสอบ", true},
		{"leading underscore", "_abc", true},
		{"trailing underscore", "abc_", true},
		{"leading hyphen", "-abc", true},
		{"trailing hyphen", "abc-", true},
		{"only separator", "_", true},
		{"too long (51)", tooLong, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateClientID(tc.input)
			if tc.wantErr && err == nil {
				t.Errorf("ValidateClientID(%q): expected error, got nil", tc.input)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ValidateClientID(%q): unexpected error: %v", tc.input, err)
			}
		})
	}
}

func TestClient_dbName(t *testing.T) {
	c := &Client{dbPrefix: "gold_shop"}

	tests := []struct {
		clientID string
		want     string
	}{
		{"000", "gold_shop"},
		{"abc", "gold_shop_abc"},
		{"Shop-01", "gold_shop_Shop-01"},
	}

	for _, tc := range tests {
		t.Run(tc.clientID, func(t *testing.T) {
			if got := c.dbName(tc.clientID); got != tc.want {
				t.Errorf("dbName(%q) = %q, want %q", tc.clientID, got, tc.want)
			}
		})
	}
}

func TestClientIDFromCtx_Roundtrip(t *testing.T) {
	ctx := WithClientID(context.Background(), "shop-42")

	got, err := ClientIDFromCtx(ctx)
	if err != nil {
		t.Fatalf("ClientIDFromCtx: unexpected error: %v", err)
	}
	if got != "shop-42" {
		t.Errorf("ClientIDFromCtx: got %q, want %q", got, "shop-42")
	}
}

func TestClientIDFromCtx_Missing(t *testing.T) {
	_, err := ClientIDFromCtx(context.Background())
	if !errors.Is(err, ErrMissingClientID) {
		t.Errorf("ClientIDFromCtx on bare ctx: got %v, want ErrMissingClientID", err)
	}
}

func TestClientIDFromCtx_EmptyTreatedAsMissing(t *testing.T) {
	ctx := WithClientID(context.Background(), "")
	_, err := ClientIDFromCtx(ctx)
	if !errors.Is(err, ErrMissingClientID) {
		t.Errorf("ClientIDFromCtx on empty: got %v, want ErrMissingClientID", err)
	}
}
