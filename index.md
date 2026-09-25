# url2svg

Capture any web page as a high-fidelity SVG using headless Chrome.

## Demo

A simple showcase page captured as SVG — text, backgrounds, borders, shadows, and links all preserved as vector elements:

<img src="demo.svg" width="600" alt="url2svg demo capture" />

## Scaling down

Because the output is vector, the same capture can be shown at any size without
re-rendering. The demo above at decreasing widths, down to icon size:

<table>
<tr>
<th>400</th><th>200</th><th>100</th><th>64</th><th>32</th><th>16</th>
</tr>
<tr style="vertical-align: top">
<td><img src="demo.svg" width="400" alt="demo at 400px" /></td>
<td><img src="demo.svg" width="200" alt="demo at 200px" /></td>
<td><img src="demo.svg" width="100" alt="demo at 100px" /></td>
<td><img src="demo.svg" width="64" alt="demo at 64px" /></td>
<td><img src="demo.svg" width="32" alt="demo at 32px" /></td>
<td><img src="demo.svg" width="16" alt="demo at 16px" /></td>
</tr>
</table>

## How it works

1. **Capture** — headless Chrome renders the page; JavaScript extracts the full visual tree
2. **Process** — the JSON tree is parsed into a Go intermediate representation
3. **Render** — the element tree is walked to emit SVG markup

## No JavaScript

The SVG is a picture, not a copy of the page's behaviour. Scripts and event handlers are never captured, `javascript:` and other non-navigation links are dropped, and inline `<svg>` elements are scrubbed of script before they are embedded. See the README for the full table.

## Links

| | |
|---|---|
| Source (Forgejo) | https://git.bytestone.uk/hum3/url2svg |
| Mirror (GitHub) | https://github.com/drummonds/url2svg |
