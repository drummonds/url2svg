package svg

import (
	"bytes"
	"fmt"
	"strings"

	"codeberg.org/hum3/url2svg/internal/model"
)

// defsCollector walks the tree and collects <clipPath> and <filter> definitions.
type defsCollector struct {
	clipPaths map[*model.VisualElement]string
	filters   map[*model.VisualElement][]string // element → list of filter IDs (one per shadow)
	nextID    int
}

func newDefsCollector() *defsCollector {
	return &defsCollector{
		clipPaths: make(map[*model.VisualElement]string),
		filters:   make(map[*model.VisualElement][]string),
	}
}

func (dc *defsCollector) genID(prefix string) string {
	dc.nextID++
	return fmt.Sprintf("%s%d", prefix, dc.nextID)
}

func (dc *defsCollector) collect(el *model.VisualElement) {
	if needsClip(el) {
		dc.clipPaths[el] = dc.genID("clip")
	}
	for range el.BoxShadows {
		id := dc.genID("shadow")
		dc.filters[el] = append(dc.filters[el], id)
	}
	for _, child := range el.Children {
		dc.collect(child)
	}
}

func needsClip(el *model.VisualElement) bool {
	return el.Overflow == "hidden" || el.Overflow == "scroll" || el.Overflow == "auto"
}

// Render converts a VisualElement tree to an SVG document.
func Render(root *model.VisualElement, viewportWidth, viewportHeight float64, mode model.Mode) ([]byte, error) {
	var buf bytes.Buffer

	precision := 2
	if mode == model.ModeCompact {
		precision = 1
	}
	w := NewWriter(&buf, precision)

	docWidth := viewportWidth
	docHeight := viewportHeight
	if docWidth == 0 {
		docWidth = 1280
	}
	if docHeight == 0 {
		docHeight = 900
	}

	dc := newDefsCollector()
	dc.collect(root)

	w.Raw(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	w.OpenSVG(docWidth, docHeight)

	// Write defs if needed
	if len(dc.clipPaths) > 0 || len(dc.filters) > 0 {
		w.OpenDefs()
		writeClipPaths(w, dc, root)
		writeFilters(w, dc, root)
		w.CloseDefs()
	}

	// White background
	w.Rect(0, 0, docWidth, docHeight, `fill="white"`)

	renderElement(w, dc, root)

	// Adaptive border — drawn last so it's on top of all content
	w.Rect(0, 0, docWidth, docHeight, fmt.Sprintf(`fill="none" stroke="%s" stroke-width="1"`, borderColor(root.BackgroundColor)))

	w.CloseSVG()
	return buf.Bytes(), nil
}

func writeClipPaths(w *Writer, dc *defsCollector, el *model.VisualElement) {
	if id, ok := dc.clipPaths[el]; ok {
		hasRadius := el.HasRadius()
		w.ClipPath(id, el.Bounds.X, el.Bounds.Y, el.Bounds.Width, el.Bounds.Height,
			hasRadius, el.BorderRadius)
	}
	for _, child := range el.Children {
		writeClipPaths(w, dc, child)
	}
}

func writeFilters(w *Writer, dc *defsCollector, el *model.VisualElement) {
	if ids, ok := dc.filters[el]; ok {
		for i, id := range ids {
			if i < len(el.BoxShadows) {
				w.Filter(id, el.BoxShadows[i])
			}
		}
	}
	for _, child := range el.Children {
		writeFilters(w, dc, child)
	}
}

func renderElement(w *Writer, dc *defsCollector, el *model.VisualElement) {
	b := el.Bounds
	if b.Width == 0 && b.Height == 0 && len(el.TextRuns) == 0 {
		return
	}

	hasOpacity := el.Opacity < 1
	hasClip := needsClip(el)
	hasLink := el.Href != ""

	// Open opacity group
	if hasOpacity {
		w.OpenGroup(fmt.Sprintf(`opacity="%s"`, w.fmtFloat(el.Opacity)))
	}

	// Open clip group
	if hasClip {
		if clipID, ok := dc.clipPaths[el]; ok {
			w.OpenGroup(fmt.Sprintf(`clip-path="url(#%s)"`, clipID))
		}
	}

	// Open link wrapper
	if hasLink {
		w.OpenLink(el.Href)
	}

	// Render box shadows (behind element)
	renderBoxShadows(w, dc, el)

	// Render background and borders
	renderBackground(w, el)

	// Render image
	if el.ImageDataURL != "" {
		w.Image(b.X, b.Y, b.Width, b.Height, el.ImageDataURL)
	}

	// Render text
	renderTextRuns(w, el)

	// Recurse children
	for _, child := range el.Children {
		renderElement(w, dc, child)
	}

	// Close wrappers in reverse order
	if hasLink {
		w.CloseLink()
	}
	if hasClip {
		if _, ok := dc.clipPaths[el]; ok {
			w.CloseGroup()
		}
	}
	if hasOpacity {
		w.CloseGroup()
	}
}

func renderBoxShadows(w *Writer, dc *defsCollector, el *model.VisualElement) {
	ids, ok := dc.filters[el]
	if !ok {
		return
	}
	b := el.Bounds
	for i, id := range ids {
		if i >= len(el.BoxShadows) || el.BoxShadows[i].Inset {
			continue // skip inset shadows for now
		}
		shadow := el.BoxShadows[i]
		// Draw a rect with the shadow filter applied
		spreadX := b.X - shadow.Spread
		spreadY := b.Y - shadow.Spread
		spreadW := b.Width + 2*shadow.Spread
		spreadH := b.Height + 2*shadow.Spread
		attrs := fmt.Sprintf(`fill="rgb(%d,%d,%d)" filter="url(#%s)"`,
			shadow.Color.R, shadow.Color.G, shadow.Color.B, id)
		if el.HasRadius() && el.IsUniformRadius() {
			w.RectRounded(spreadX, spreadY, spreadW, spreadH, el.BorderRadius.TopLeft, attrs)
		} else {
			w.Rect(spreadX, spreadY, spreadW, spreadH, attrs)
		}
	}
}

func renderBackground(w *Writer, el *model.VisualElement) {
	b := el.Bounds
	bg := el.BackgroundColor
	hasBG := bg.A > 0
	hasBorder := el.HasBorder()

	if !hasBG && !hasBorder {
		return
	}

	var attrParts []string

	// Fill
	if hasBG {
		color, opacity := w.fmtColor(bg)
		attrParts = append(attrParts, fmt.Sprintf(`fill="%s"`, color))
		if bg.A < 1 {
			attrParts = append(attrParts, fmt.Sprintf(`fill-opacity="%s"`, opacity))
		}
	} else {
		attrParts = append(attrParts, `fill="none"`)
	}

	// Uniform border → stroke on the rect
	if hasBorder && el.IsUniformBorder() {
		border := el.Borders[0]
		if border.Width > 0 && border.Style != "none" {
			color, opacity := w.fmtColor(border.Color)
			attrParts = append(attrParts, fmt.Sprintf(`stroke="%s"`, color))
			attrParts = append(attrParts, fmt.Sprintf(`stroke-width="%s"`, w.fmtFloat(border.Width)))
			if border.Color.A < 1 {
				attrParts = append(attrParts, fmt.Sprintf(`stroke-opacity="%s"`, opacity))
			}
			switch border.Style {
			case "dashed":
				attrParts = append(attrParts, fmt.Sprintf(`stroke-dasharray="%s %s"`,
					w.fmtFloat(border.Width*3), w.fmtFloat(border.Width*3)))
			case "dotted":
				attrParts = append(attrParts, fmt.Sprintf(`stroke-dasharray="%s %s"`,
					w.fmtFloat(border.Width), w.fmtFloat(border.Width)))
			}
		}
	}

	attrs := strings.Join(attrParts, " ")

	if el.HasRadius() {
		if el.IsUniformRadius() {
			w.RectRounded(b.X, b.Y, b.Width, b.Height, el.BorderRadius.TopLeft, attrs)
		} else {
			d := w.roundedRectPath(b.X, b.Y, b.Width, b.Height, el.BorderRadius)
			w.Path(d, attrs)
		}
	} else {
		w.Rect(b.X, b.Y, b.Width, b.Height, attrs)
	}

	// Non-uniform borders → individual lines
	if hasBorder && !el.IsUniformBorder() {
		renderNonUniformBorders(w, el)
	}
}

func renderNonUniformBorders(w *Writer, el *model.VisualElement) {
	b := el.Bounds
	sides := [4]struct {
		border model.Border
		x1, y1 float64
		x2, y2 float64
	}{
		{el.Borders[0], b.X, b.Y, b.X + b.Width, b.Y},                       // top
		{el.Borders[1], b.X + b.Width, b.Y, b.X + b.Width, b.Y + b.Height},  // right
		{el.Borders[2], b.X, b.Y + b.Height, b.X + b.Width, b.Y + b.Height}, // bottom
		{el.Borders[3], b.X, b.Y, b.X, b.Y + b.Height},                      // left
	}

	for _, s := range sides {
		if s.border.Width <= 0 || s.border.Style == "none" || s.border.Style == "" {
			continue
		}
		color, opacity := w.fmtColor(s.border.Color)
		var attrParts []string
		attrParts = append(attrParts, fmt.Sprintf(`stroke="%s"`, color))
		attrParts = append(attrParts, fmt.Sprintf(`stroke-width="%s"`, w.fmtFloat(s.border.Width)))
		if s.border.Color.A < 1 {
			attrParts = append(attrParts, fmt.Sprintf(`stroke-opacity="%s"`, opacity))
		}
		switch s.border.Style {
		case "dashed":
			attrParts = append(attrParts, fmt.Sprintf(`stroke-dasharray="%s %s"`,
				w.fmtFloat(s.border.Width*3), w.fmtFloat(s.border.Width*3)))
		case "dotted":
			attrParts = append(attrParts, fmt.Sprintf(`stroke-dasharray="%s %s"`,
				w.fmtFloat(s.border.Width), w.fmtFloat(s.border.Width)))
		}
		w.Line(s.x1, s.y1, s.x2, s.y2, strings.Join(attrParts, " "))
	}
}

func renderTextRuns(w *Writer, el *model.VisualElement) {
	if len(el.TextRuns) == 0 {
		return
	}

	color, opacity := w.fmtColor(el.Color)

	for _, run := range el.TextRuns {
		if run.Text == "" {
			continue
		}

		// Approximate baseline: top + 80% of line height
		baselineY := run.Bounds.Y + run.Bounds.Height*0.8

		var attrParts []string
		attrParts = append(attrParts, fmt.Sprintf(`fill="%s"`, color))
		if el.Color.A < 1 {
			attrParts = append(attrParts, fmt.Sprintf(`fill-opacity="%s"`, opacity))
		}
		if el.FontFamily != "" {
			attrParts = append(attrParts, fmt.Sprintf(`font-family="%s"`, xmlEscape(el.FontFamily)))
		}
		if el.FontSize > 0 {
			attrParts = append(attrParts, fmt.Sprintf(`font-size="%s"`, w.fmtFloat(el.FontSize)))
		}
		if el.FontWeight != "" && el.FontWeight != "400" && el.FontWeight != "normal" {
			attrParts = append(attrParts, fmt.Sprintf(`font-weight="%s"`, el.FontWeight))
		}
		if el.FontStyle != "" && el.FontStyle != "normal" {
			attrParts = append(attrParts, fmt.Sprintf(`font-style="%s"`, el.FontStyle))
		}
		if el.TextDecoration != "" && el.TextDecoration != "none" {
			attrParts = append(attrParts, fmt.Sprintf(`text-decoration="%s"`, el.TextDecoration))
		}

		w.Text(run.Bounds.X, baselineY, run.Text, strings.Join(attrParts, " "))
	}
}

// borderColor returns a border color that contrasts with the effective background.
// It alpha-blends bg over white, then checks luminance.
func borderColor(bg model.Color) string {
	// Alpha-blend over white
	r := float64(bg.R)*bg.A + 255*(1-bg.A)
	g := float64(bg.G)*bg.A + 255*(1-bg.A)
	b := float64(bg.B)*bg.A + 255*(1-bg.A)

	// Relative luminance (ITU-R BT.709)
	luminance := (0.2126*r + 0.7152*g + 0.0722*b) / 255

	if luminance > 0.5 {
		return "#ccc"
	}
	return "#555"
}
