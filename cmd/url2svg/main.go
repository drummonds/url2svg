package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"codeberg.org/hum3/url2svg/internal/capture"
	"codeberg.org/hum3/url2svg/internal/model"
	"codeberg.org/hum3/url2svg/internal/process"
	"codeberg.org/hum3/url2svg/internal/svg"
)

func main() {
	url := flag.String("url", "", "URL to capture (required)")
	output := flag.String("o", "output.svg", "Output file path")
	width := flag.Int("width", 0, "Viewport width (default: 1024 compact, 1280 full)")
	height := flag.Int("height", 0, "Viewport height (default: 4:3 ratio of width)")
	timeout := flag.Int("timeout", 30, "Page load timeout in seconds")
	fullPage := flag.Bool("full-page", false, "Capture full scroll height")
	modeFlag := flag.String("mode", "compact", "Output mode: compact or full")
	flag.Parse()

	if *url == "" {
		fmt.Fprintln(os.Stderr, "error: -url is required")
		flag.Usage()
		os.Exit(1)
	}

	mode := model.Mode(*modeFlag)
	if mode != model.ModeFull && mode != model.ModeCompact {
		fmt.Fprintf(os.Stderr, "error: -mode must be \"compact\" or \"full\", got %q\n", *modeFlag)
		os.Exit(1)
	}

	// Mode-dependent viewport defaults
	if *width == 0 {
		if mode == model.ModeCompact {
			*width = 1024
		} else {
			*width = 1280
		}
	}
	if *height == 0 {
		*height = *width * 3 / 4
	}

	fmt.Fprintf(os.Stderr, "Capturing %s ...\n", *url)

	jsonStr, err := capture.Capture(context.Background(), capture.Options{
		URL:      *url,
		Width:    *width,
		Height:   *height,
		Timeout:  time.Duration(*timeout) * time.Second,
		FullPage: *fullPage,
		Mode:     mode,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: capture failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Processing visual tree (%d bytes) ...\n", len(jsonStr))

	result, err := process.Parse(jsonStr, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: processing failed: %v\n", err)
		os.Exit(1)
	}

	// Use configured viewport height unless full-page captures the scroll height
	renderWidth := result.ViewportWidth
	renderHeight := float64(*height)
	if *fullPage {
		renderHeight = result.ViewportHeight
	}

	fmt.Fprintln(os.Stderr, "Rendering SVG ...")

	svgData, err := svg.Render(result.Root, renderWidth, renderHeight, mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: rendering failed: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, svgData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Done! Written to %s (%d bytes)\n", *output, len(svgData))
}
