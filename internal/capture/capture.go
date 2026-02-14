package capture

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/hum3/url2svg/internal/model"
)

//go:embed extract.js
var extractJS string

// Options configures the browser capture.
type Options struct {
	URL      string
	Width    int
	Height   int
	Timeout  time.Duration
	FullPage bool
	Mode     model.Mode
}

// Capture navigates to the URL and extracts the visual tree as a JSON string.
func Capture(ctx context.Context, opts Options) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	path, found := launcher.LookPath()
	if !found {
		return "", fmt.Errorf("Chrome/Chromium not found in PATH")
	}

	u := launcher.New().Bin(path).Headless(true).MustLaunch()

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return "", fmt.Errorf("connecting to browser: %w", err)
	}
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("creating page: %w", err)
	}

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  opts.Width,
		Height: opts.Height,
	}); err != nil {
		return "", fmt.Errorf("setting viewport: %w", err)
	}

	if err := page.Context(ctx).Navigate(opts.URL); err != nil {
		return "", fmt.Errorf("navigating to %s: %w", opts.URL, err)
	}

	if err := page.Context(ctx).WaitLoad(); err != nil {
		return "", fmt.Errorf("waiting for page load: %w", err)
	}

	// Give JS frameworks a moment to render
	time.Sleep(500 * time.Millisecond)

	if opts.FullPage {
		// Scroll to bottom to trigger lazy loading, then back to top
		page.MustEval(`() => window.scrollTo(0, document.body.scrollHeight)`)
		time.Sleep(300 * time.Millisecond)
		page.MustEval(`() => window.scrollTo(0, 0)`)
		time.Sleep(200 * time.Millisecond)
	}

	jsOpts := map[string]any{
		"imageMode":   string(opts.Mode),
		"jpegQuality": 0.7,
	}
	result, err := page.Eval(extractJS, jsOpts)
	if err != nil {
		return "", fmt.Errorf("evaluating extraction script: %w", err)
	}

	jsonStr := result.Value.Str()
	if jsonStr == "" {
		return "", fmt.Errorf("extraction returned empty result")
	}

	return jsonStr, nil
}
