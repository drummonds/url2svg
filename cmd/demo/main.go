// Command demo serves a simple showcase page for url2svg capture demos.
package main

import (
	"fmt"
	"log"
	"net/http"
)

const page = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>url2svg Demo</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; color: #333; padding: 32px; }
  h1 { font-size: 28px; margin-bottom: 8px; }
  .subtitle { color: #666; font-size: 16px; margin-bottom: 24px; }
  .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; max-width: 900px; }
  .card { background: #fff; border-radius: 8px; padding: 20px; border: 1px solid #e0e0e0; }
  .card h2 { font-size: 16px; margin-bottom: 8px; color: #222; }
  .card p { font-size: 14px; line-height: 1.5; color: #555; }
  .card-shadow { box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
  .card-blue { background: #e3f2fd; border-color: #90caf9; }
  .card-green { background: #e8f5e9; border-color: #a5d6a7; }
  .badge { display: inline-block; background: #1976d2; color: #fff; font-size: 12px; padding: 2px 8px; border-radius: 12px; margin-right: 6px; }
  .badge-green { background: #388e3c; }
  a { color: #1565c0; text-decoration: underline; }
  .footer { margin-top: 24px; font-size: 12px; color: #999; }
</style>
</head>
<body>
  <h1>url2svg Demo Page</h1>
  <p class="subtitle">A showcase of elements captured as SVG</p>
  <div class="grid">
    <div class="card card-shadow">
      <h2>Text &amp; Typography</h2>
      <p>Regular text, <strong>bold text</strong>, <em>italic text</em>, and <a href="https://codeberg.org/hum3/url2svg">a link</a>.</p>
    </div>
    <div class="card card-blue">
      <h2>Coloured Background</h2>
      <p><span class="badge">SVG</span><span class="badge badge-green">Vector</span></p>
      <p>Backgrounds, borders, and inline badges with rounded corners.</p>
    </div>
    <div class="card card-green">
      <h2>Border Radius</h2>
      <p>Cards with rounded corners and coloured borders demonstrate border rendering fidelity.</p>
    </div>
    <div class="card card-shadow">
      <h2>Box Shadows</h2>
      <p>This card has a box shadow rendered via SVG filter primitives.</p>
    </div>
  </div>
  <p class="footer">Captured by url2svg &mdash; <a href="https://codeberg.org/hum3/url2svg">codeberg.org/hum3/url2svg</a></p>
</body>
</html>`

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, page)
	})
	log.Println("demo server listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
