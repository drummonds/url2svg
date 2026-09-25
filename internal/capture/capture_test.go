package capture

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"codeberg.org/hum3/url2svg/internal/model"
	"github.com/go-rod/rod/lib/launcher"
)

// An inline <svg> is captured as one image at its rendered size. Walking into
// its children would emit each shape as a separate image and duplicate every
// <text> as an HTML text run at the unscaled CSS font size.
func TestCapture_InlineSVGIsASingleImage(t *testing.T) {
	if _, found := launcher.LookPath(); !found {
		t.Skip("Chrome/Chromium not found")
	}

	const page = `<!DOCTYPE html><html><body>
<h1>Heading</h1>
<div style="width:400px">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 400" style="width:100%">
  <rect x="0" y="0" width="800" height="400" fill="#eee"/>
  <line x1="10" y1="10" x2="790" y2="390" stroke="#333"/>
  <text x="100" y="200" font-size="12">tick label</text>
</svg>
</div>
</body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(page))
	}))
	defer srv.Close()

	raw, err := Capture(context.Background(), Options{
		URL: srv.URL, Width: 1024, Height: 768, Timeout: 30 * time.Second, Mode: model.ModeFull,
	})
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Root *model.VisualElement `json:"root"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}

	var svgs []*model.VisualElement
	var texts int
	var walk func(el *model.VisualElement)
	walk = func(el *model.VisualElement) {
		switch el.Tag {
		case "svg":
			svgs = append(svgs, el)
		case "text", "line", "rect":
			texts++
		}
		for _, c := range el.Children {
			walk(c)
		}
	}
	walk(doc.Root)

	if len(svgs) != 1 {
		t.Fatalf("expected exactly one svg node, got %d", len(svgs))
	}
	svg := svgs[0]
	if !strings.HasPrefix(svg.ImageDataURL, "data:image/svg+xml") {
		t.Errorf("svg node should carry the serialised SVG as its image, got %q", svg.ImageDataURL[:min(40, len(svg.ImageDataURL))])
	}
	if len(svg.Children) != 0 {
		t.Errorf("svg node should be a leaf, got %d children", len(svg.Children))
	}
	if len(svg.TextRuns) != 0 {
		t.Errorf("svg node should have no HTML text runs, got %d", len(svg.TextRuns))
	}
	if texts != 0 {
		t.Errorf("svg shapes and text should not appear as separate nodes, got %d", texts)
	}
	if svg.Bounds.Width < 390 || svg.Bounds.Width > 410 {
		t.Errorf("svg should be laid out at its rendered width (~400), got %v", svg.Bounds.Width)
	}
}
