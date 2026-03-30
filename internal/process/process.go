package process

import (
	"encoding/json"
	"fmt"

	"codeberg.org/hum3/url2svg/internal/model"
)

// extractionResult wraps the JS output which includes viewport dimensions and the tree.
type extractionResult struct {
	Viewport struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	} `json:"viewport"`
	Root *model.VisualElement `json:"root"`
}

// Result holds the parsed visual tree and page dimensions.
type Result struct {
	Root           *model.VisualElement
	ViewportWidth  float64
	ViewportHeight float64
}

// Parse decodes the JSON string from JS extraction into a VisualElement tree.
func Parse(jsonStr string, mode model.Mode) (*Result, error) {
	var raw extractionResult
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, fmt.Errorf("parsing visual tree JSON: %w", err)
	}
	if raw.Root == nil {
		return nil, fmt.Errorf("extraction returned null root element")
	}
	normalize(raw.Root)
	if mode == model.ModeCompact {
		pruneEmpty(raw.Root)
	}
	return &Result{
		Root:           raw.Root,
		ViewportWidth:  raw.Viewport.Width,
		ViewportHeight: raw.Viewport.Height,
	}, nil
}

// normalize walks the tree and fixes up any values that need post-processing.
func normalize(el *model.VisualElement) {
	// Clamp opacity to [0, 1]
	if el.Opacity < 0 {
		el.Opacity = 0
	}
	if el.Opacity > 1 {
		el.Opacity = 1
	}

	// Clamp color alpha values
	clampAlpha(&el.BackgroundColor)
	clampAlpha(&el.Color)
	for i := range el.Borders {
		clampAlpha(&el.Borders[i].Color)
	}
	for i := range el.BoxShadows {
		clampAlpha(&el.BoxShadows[i].Color)
	}

	for _, child := range el.Children {
		normalize(child)
	}
}

// pruneEmpty recursively removes leaf elements that have no visual content.
func pruneEmpty(el *model.VisualElement) {
	// Recurse first so we prune bottom-up
	kept := el.Children[:0]
	for _, child := range el.Children {
		pruneEmpty(child)
		if !isEmptyLeaf(child) {
			kept = append(kept, child)
		}
	}
	el.Children = kept
}

// isEmptyLeaf returns true if the element is a leaf with no visible content.
func isEmptyLeaf(el *model.VisualElement) bool {
	if len(el.Children) > 0 {
		return false
	}
	if el.ImageDataURL != "" {
		return false
	}
	if len(el.TextRuns) > 0 {
		return false
	}
	if el.BackgroundColor.A > 0 {
		return false
	}
	for _, b := range el.Borders {
		if b.Width > 0 && b.Style != "none" && b.Style != "" {
			return false
		}
	}
	return len(el.BoxShadows) == 0
}

func clampAlpha(c *model.Color) {
	if c.A < 0 {
		c.A = 0
	}
	if c.A > 1 {
		c.A = 1
	}
}
