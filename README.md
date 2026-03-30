# url2svg

Capture any web page as a high-fidelity SVG using headless Chrome.

url2svg renders a URL in a real browser, extracts the full visual tree (elements, text, images, styles), and produces a standalone SVG file — no screenshots, no rasterisation. The output is a scalable vector representation of what the page actually looks like.

## Example

```bash
url2svg -url https://example.com -o page.svg
```

See the [documentation](https://h3-url2svg.statichost.page/) for a live demo capture.

## Features

- Backgrounds, borders, border-radius, opacity, overflow clipping
- Text runs with correct font family, size, weight, style, and decoration
- Images embedded as inline data URLs (no external dependencies)
- Box shadows
- Clickable links preserved via `<a>` elements
- Full-page scroll capture for below-the-fold content
- Compact mode for smaller file sizes

## Installation

Requires Go 1.25+ and Chrome/Chromium installed on the system.

```bash
go install codeberg.org/hum3/url2svg/cmd/url2svg@latest
```

Or build from source:

```bash
git clone https://codeberg.org/hum3/url2svg.git
cd url2svg
go build -o url2svg ./cmd/url2svg
```

## Usage

```
url2svg -url <URL> [options]
```

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-url` | *(required)* | URL to capture |
| `-o` | `output.svg` | Output file path |
| `-mode` | `compact` | Output mode: `compact` or `full` |
| `-width` | `1024` (compact) / `1280` (full) | Viewport width in pixels |
| `-height` | 4:3 ratio of width | Viewport height in pixels |
| `-full-page` | `false` | Capture the full scroll height, not just the viewport |
| `-timeout` | `30` | Page load timeout in seconds |

### Output modes

**compact** (default) — Optimised for smaller files. Uses a 1024px viewport, downscales embedded images, and rounds coordinates to reduce SVG size.

**full** — Maximum fidelity. Uses a 1280px viewport and preserves images at original quality. Produces larger files.

### Examples

Basic capture with defaults (compact mode, 1024x768):

```bash
url2svg -url https://example.com
```

Full-fidelity capture at a custom resolution:

```bash
url2svg -url https://example.com -mode full -width 1920 -height 1080 -o page.svg
```

Capture the entire page including below-the-fold content:

```bash
url2svg -url https://example.com -full-page -o full-page.svg
```

Set a longer timeout for slow-loading pages:

```bash
url2svg -url https://example.com -timeout 60
```

## How it works

1. **Capture** — Launches headless Chrome via [go-rod](https://github.com/nicedoc/go-rod), navigates to the URL, waits for the page to load, then runs a JavaScript extraction script that walks the DOM and collects every visible element's geometry, styles, text content, and embedded images.

2. **Process** — Parses the JSON visual tree into a Go intermediate representation (`VisualElement` tree), applying mode-specific optimisations.

3. **Render** — Walks the element tree and emits SVG markup: rectangles for boxes, `<text>` for text runs, `<image>` for embedded images, `<clipPath>` for overflow clipping, and `<filter>` for box shadows.

## Requirements

- Go 1.25+
- Chrome or Chromium (detected automatically via `$PATH`)

## License

MIT

## Links

| | |
|---|---|
| Documentation | https://h3-url2svg.statichost.page/ |
| Source (Codeberg) | https://codeberg.org/hum3/url2svg |
| Mirror (GitHub) | https://github.com/drummonds/url2svg |
