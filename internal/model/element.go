package model

// Mode controls output fidelity.
type Mode string

const (
	ModeFull    Mode = "full"
	ModeCompact Mode = "compact"
)

// Rect represents a bounding rectangle with absolute page coordinates.
type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Color represents a pre-parsed RGBA color.
type Color struct {
	R int     `json:"r"`
	G int     `json:"g"`
	B int     `json:"b"`
	A float64 `json:"a"`
}

// Border represents a single border side.
type Border struct {
	Width float64 `json:"width"`
	Style string  `json:"style"`
	Color Color   `json:"color"`
}

// BorderRadius holds the four corner radii.
type BorderRadius struct {
	TopLeft     float64 `json:"topLeft"`
	TopRight    float64 `json:"topRight"`
	BottomRight float64 `json:"bottomRight"`
	BottomLeft  float64 `json:"bottomLeft"`
}

// Shadow represents a box-shadow.
type Shadow struct {
	OffsetX float64 `json:"offsetX"`
	OffsetY float64 `json:"offsetY"`
	Blur    float64 `json:"blur"`
	Spread  float64 `json:"spread"`
	Color   Color   `json:"color"`
	Inset   bool    `json:"inset"`
}

// TextRun represents a single line of text with its bounding box.
type TextRun struct {
	Text   string `json:"text"`
	Bounds Rect   `json:"bounds"`
}

// VisualElement is the core intermediate representation node.
type VisualElement struct {
	// Geometry
	Bounds Rect `json:"bounds"`

	// Visual properties
	BackgroundColor Color        `json:"backgroundColor"`
	Borders         [4]Border    `json:"borders"` // top, right, bottom, left
	BorderRadius    BorderRadius `json:"borderRadius"`
	Opacity         float64      `json:"opacity"`
	Overflow        string       `json:"overflow"`
	BoxShadows      []Shadow     `json:"boxShadows"`

	// Text properties
	FontFamily     string  `json:"fontFamily"`
	FontSize       float64 `json:"fontSize"`
	FontWeight     string  `json:"fontWeight"`
	FontStyle      string  `json:"fontStyle"`
	Color          Color   `json:"color"`
	TextDecoration string  `json:"textDecoration"`
	LineHeight     float64 `json:"lineHeight"`
	TextAlign      string  `json:"textAlign"`

	// Semantic properties
	Tag       string `json:"tag"`
	ID        string `json:"id"`
	Classes   string `json:"classes"`
	Role      string `json:"role"`
	AriaLabel string `json:"ariaLabel"`
	Href      string `json:"href"`

	// Content
	ImageDataURL string           `json:"imageDataURL"`
	TextRuns     []TextRun        `json:"textRuns"`
	Children     []*VisualElement `json:"children"`
}

// IsUniformBorder returns true if all four borders have the same width, style, and color.
func (e *VisualElement) IsUniformBorder() bool {
	b := e.Borders
	return b[0].Width == b[1].Width && b[1].Width == b[2].Width && b[2].Width == b[3].Width &&
		b[0].Style == b[1].Style && b[1].Style == b[2].Style && b[2].Style == b[3].Style &&
		b[0].Color == b[1].Color && b[1].Color == b[2].Color && b[2].Color == b[3].Color
}

// HasBorder returns true if any border side has a visible width.
func (e *VisualElement) HasBorder() bool {
	for _, b := range e.Borders {
		if b.Width > 0 && b.Style != "none" && b.Style != "" {
			return true
		}
	}
	return false
}

// IsUniformRadius returns true if all four corner radii are equal.
func (e *VisualElement) IsUniformRadius() bool {
	r := e.BorderRadius
	return r.TopLeft == r.TopRight && r.TopRight == r.BottomRight && r.BottomRight == r.BottomLeft
}

// HasRadius returns true if any corner radius is non-zero.
func (e *VisualElement) HasRadius() bool {
	r := e.BorderRadius
	return r.TopLeft > 0 || r.TopRight > 0 || r.BottomRight > 0 || r.BottomLeft > 0
}
