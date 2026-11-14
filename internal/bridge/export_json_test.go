package bridge

import (
	"strings"
	"testing"
)

func TestMarshalJSONKeepsEntriesWhenAFieldIsNotUTF8(t *testing.T) {
	type entry struct {
		Title    string `json:"title"`
		Password string `json:"password"`
		Notes    string `json:"notes"`
	}
	got := string(marshalJSON([]entry{
		{Title: "Bank", Password: "p&<>", Notes: "ok"},
		{Title: "Other", Password: "a\xffb"},
	}))
	if !strings.HasPrefix(got, "[") || !strings.HasSuffix(got, "]") {
		t.Fatalf("truncated JSON: %q", got)
	}
	replaced := strings.Contains(got, "a\uFFFDb") || strings.Contains(got, `a\ufffdb`)
	if strings.Contains(got, "\xff") || !strings.Contains(got, `"password":"p&<>"`) || !replaced {
		t.Fatalf("invalid UTF-8 was not replaced with U+FFFD: %s", got)
	}
	if !strings.Contains(got, `"title":"Bank"`) || !strings.Contains(got, `"title":"Other"`) {
		t.Fatalf("missing entries: %s", got)
	}
}

func TestMarshalJSONNilSliceIsEmptyArray(t *testing.T) {
	var groups []string
	if got := string(marshalJSON(groups)); got != "[]" {
		t.Fatalf("got %q", got)
	}
}
