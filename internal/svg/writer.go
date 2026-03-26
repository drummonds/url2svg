package svg

import (
	"fmt"
	"io"
	"math"
	"strings"

	"codeberg.org/hum3/url2svg/internal/model"
)

// Writer provides low-level SVG element writing with float64 precision.
type Writer struct {
	w         io.Writer
	indent    int
	Precision int
}

// NewWriter creates a new SVG Writer with the given coordinate precision.
func NewWriter(w io.Writer, precision int) *Writer {
	return &Writer{w: w, Precision: precision}
}

// fmtFloat formats a float64 with trailing zero trimming, using w.Precision decimal places.
func (w *Writer) fmtFloat(f float64) string {
	if f == 0 {
		return "0"
	}
	s := fmt.Sprintf("%.*f", w.Precision, f)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// fmtColor formats a Color as an SVG fill/stroke value. Returns the color string and opacity.
func (w *Writer) fmtColor(c model.Color) (string, string) {
	color := fmt.Sprintf("rgb(%d,%d,%d)", c.R, c.G, c.B)
	opacity := w.fmtFloat(c.A)
	return color, opacity
}

func (w *Writer) writeIndent() {
	for i := 0; i < w.indent; i++ {
		fmt.Fprint(w.w, "  ")
	}
}

// Indent increases the indentation level.
func (w *Writer) Indent() { w.indent++ }

// Dedent decreases the indentation level.
func (w *Writer) Dedent() {
	if w.indent > 0 {
		w.indent--
	}
}

// Raw writes a raw string.
func (w *Writer) Raw(s string) {
	fmt.Fprint(w.w, s)
}

// Linef writes an indented formatted line.
func (w *Writer) Linef(format string, args ...any) {
	w.writeIndent()
	fmt.Fprintf(w.w, format, args...)
	fmt.Fprint(w.w, "\n")
}

// OpenSVG writes the opening <svg> tag with the given dimensions.
func (w *Writer) OpenSVG(width, height float64) {
	w.Linef(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%s" height="%s" viewBox="0 0 %s %s">`,
		w.fmtFloat(width), w.fmtFloat(height), w.fmtFloat(width), w.fmtFloat(height))
	w.Indent()
}

// CloseSVG writes the closing </svg> tag.
func (w *Writer) CloseSVG() {
	w.Dedent()
	w.Linef(`</svg>`)
}

// OpenDefs writes <defs>.
func (w *Writer) OpenDefs() {
	w.Linef(`<defs>`)
	w.Indent()
}

// CloseDefs writes </defs>.
func (w *Writer) CloseDefs() {
	w.Dedent()
	w.Linef(`</defs>`)
}

// OpenGroup writes an opening <g> tag with optional attributes.
func (w *Writer) OpenGroup(attrs string) {
	if attrs != "" {
		w.Linef(`<g %s>`, attrs)
	} else {
		w.Linef(`<g>`)
	}
	w.Indent()
}

// CloseGroup writes </g>.
func (w *Writer) CloseGroup() {
	w.Dedent()
	w.Linef(`</g>`)
}

// OpenLink writes an opening <a> tag.
func (w *Writer) OpenLink(href string) {
	w.Linef(`<a xlink:href="%s">`, xmlEscape(href))
	w.Indent()
}

// CloseLink writes </a>.
func (w *Writer) CloseLink() {
	w.Dedent()
	w.Linef(`</a>`)
}

// Rect writes a <rect> element.
func (w *Writer) Rect(x, y, width, height float64, attrs string) {
	w.Linef(`<rect x="%s" y="%s" width="%s" height="%s" %s/>`,
		w.fmtFloat(x), w.fmtFloat(y), w.fmtFloat(width), w.fmtFloat(height), attrs)
}

// RectRounded writes a <rect> with uniform border radius.
func (w *Writer) RectRounded(x, y, width, height, rx float64, attrs string) {
	w.Linef(`<rect x="%s" y="%s" width="%s" height="%s" rx="%s" %s/>`,
		w.fmtFloat(x), w.fmtFloat(y), w.fmtFloat(width), w.fmtFloat(height), w.fmtFloat(rx), attrs)
}

// Path writes a <path> element.
func (w *Writer) Path(d string, attrs string) {
	w.Linef(`<path d="%s" %s/>`, d, attrs)
}

// Line writes a <line> element.
func (w *Writer) Line(x1, y1, x2, y2 float64, attrs string) {
	w.Linef(`<line x1="%s" y1="%s" x2="%s" y2="%s" %s/>`,
		w.fmtFloat(x1), w.fmtFloat(y1), w.fmtFloat(x2), w.fmtFloat(y2), attrs)
}

// Text writes a <text> element.
func (w *Writer) Text(x, y float64, content string, attrs string) {
	w.Linef(`<text x="%s" y="%s" %s>%s</text>`,
		w.fmtFloat(x), w.fmtFloat(y), attrs, xmlEscape(content))
}

// Image writes an <image> element with a data URL.
func (w *Writer) Image(x, y, width, height float64, dataURL string) {
	w.Linef(`<image x="%s" y="%s" width="%s" height="%s" href="%s"/>`,
		w.fmtFloat(x), w.fmtFloat(y), w.fmtFloat(width), w.fmtFloat(height), xmlEscape(dataURL))
}

// ClipPath writes a <clipPath> definition with a rect.
func (w *Writer) ClipPath(id string, x, y, width, height float64, hasRadius bool, radius model.BorderRadius) {
	w.Linef(`<clipPath id="%s">`, id)
	w.Indent()
	if hasRadius && !isUniformRadius(radius) {
		// Non-uniform radius: use path
		d := w.roundedRectPath(x, y, width, height, radius)
		w.Linef(`<path d="%s"/>`, d)
	} else if hasRadius {
		w.RectRounded(x, y, width, height, radius.TopLeft, "")
	} else {
		w.Rect(x, y, width, height, "")
	}
	w.Dedent()
	w.Linef(`</clipPath>`)
}

// Filter writes an SVG filter for a box shadow.
func (w *Writer) Filter(id string, shadow model.Shadow) {
	// Expand filter region to accommodate blur and offset
	expand := shadow.Blur + math.Abs(shadow.Spread)
	fx := -expand/100 - 0.5
	fy := -expand/100 - 0.5
	fw := 1 + expand/50 + 1
	fh := 1 + expand/50 + 1
	w.Linef(`<filter id="%s" x="%s" y="%s" width="%s" height="%s">`,
		id, w.fmtFloat(fx), w.fmtFloat(fy), w.fmtFloat(fw), w.fmtFloat(fh))
	w.Indent()
	stdDev := shadow.Blur / 2
	w.Linef(`<feGaussianBlur in="SourceAlpha" stdDeviation="%s"/>`, w.fmtFloat(stdDev))
	w.Linef(`<feOffset dx="%s" dy="%s" result="offsetblur"/>`, w.fmtFloat(shadow.OffsetX), w.fmtFloat(shadow.OffsetY))
	color, _ := w.fmtColor(shadow.Color)
	w.Linef(`<feFlood flood-color="%s" flood-opacity="%s"/>`, color, w.fmtFloat(shadow.Color.A))
	w.Linef(`<feComposite in2="offsetblur" operator="in"/>`)
	w.Linef(`<feMerge>`)
	w.Indent()
	w.Linef(`<feMergeNode/>`)
	w.Linef(`<feMergeNode in="SourceGraphic"/>`)
	w.Dedent()
	w.Linef(`</feMerge>`)
	w.Dedent()
	w.Linef(`</filter>`)
}

func isUniformRadius(r model.BorderRadius) bool {
	return r.TopLeft == r.TopRight && r.TopRight == r.BottomRight && r.BottomRight == r.BottomLeft
}

// roundedRectPath generates an SVG path for a rectangle with non-uniform border radii.
func (wr *Writer) roundedRectPath(x, y, w, h float64, r model.BorderRadius) string {
	tl := math.Min(r.TopLeft, math.Min(w/2, h/2))
	tr := math.Min(r.TopRight, math.Min(w/2, h/2))
	br := math.Min(r.BottomRight, math.Min(w/2, h/2))
	bl := math.Min(r.BottomLeft, math.Min(w/2, h/2))

	return fmt.Sprintf("M%s,%s H%s Q%s,%s %s,%s V%s Q%s,%s %s,%s H%s Q%s,%s %s,%s V%s Q%s,%s %s,%sZ",
		wr.fmtFloat(x+tl), wr.fmtFloat(y),
		wr.fmtFloat(x+w-tr),
		wr.fmtFloat(x+w), wr.fmtFloat(y), wr.fmtFloat(x+w), wr.fmtFloat(y+tr),
		wr.fmtFloat(y+h-br),
		wr.fmtFloat(x+w), wr.fmtFloat(y+h), wr.fmtFloat(x+w-br), wr.fmtFloat(y+h),
		wr.fmtFloat(x+bl),
		wr.fmtFloat(x), wr.fmtFloat(y+h), wr.fmtFloat(x), wr.fmtFloat(y+h-bl),
		wr.fmtFloat(y+tl),
		wr.fmtFloat(x), wr.fmtFloat(y), wr.fmtFloat(x+tl), wr.fmtFloat(y),
	)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
