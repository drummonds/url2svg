# url2svg

Capture any web page as a high-fidelity SVG using headless Chrome.

## Demo

A simple showcase page captured as SVG — text, backgrounds, borders, shadows, and links all preserved as vector elements:

<img src="demo.svg" width="600" alt="url2svg demo capture" />

## How it works

1. **Capture** — headless Chrome renders the page; JavaScript extracts the full visual tree
2. **Process** — the JSON tree is parsed into a Go intermediate representation
3. **Render** — the element tree is walked to emit SVG markup

## Links

| | |
|---|---|
| Source (Codeberg) | https://codeberg.org/hum3/url2svg |
| Mirror (GitHub) | https://github.com/drummonds/url2svg |
