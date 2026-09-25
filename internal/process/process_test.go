package process

import (
	"fmt"
	"testing"

	"git.bytestone.uk/hum3/url2svg/internal/model"
)

// Only web and mail links survive into the SVG; anything that could run
// code when clicked (javascript:, data:, vbscript:) is dropped.
func TestParse_SanitisesLinkSchemes(t *testing.T) {
	cases := []struct{ href, want string }{
		{"https://example.com/a", "https://example.com/a"},
		{"http://example.com/a", "http://example.com/a"},
		{"mailto:someone@example.com", "mailto:someone@example.com"},
		{"javascript:alert(1)", ""},
		{"JavaScript:alert(1)", ""},
		{" javascript:alert(1)", ""},
		{"data:text/html,<script>alert(1)</script>", ""},
		{"vbscript:msgbox", ""},
		{"", ""},
	}
	for _, c := range cases {
		jsonStr := fmt.Sprintf(`{"viewport":{"width":10,"height":10},"root":{"tag":"a","href":%q,"opacity":1,"children":[]}}`, c.href)
		res, err := Parse(jsonStr, model.ModeFull)
		if err != nil {
			t.Fatalf("href %q: %v", c.href, err)
		}
		if res.Root.Href != c.want {
			t.Errorf("href %q: got %q, want %q", c.href, res.Root.Href, c.want)
		}
	}
}
