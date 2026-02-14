# url2svg Roadmap

## Phase 1 — Core capture (done)
- URL → headless Chrome → JS extraction → VisualElement tree → SVG
- Backgrounds, borders, border-radius, opacity, overflow clipping
- Text runs with font properties, text decoration
- Images as embedded data URLs
- Box shadows
- Links (href)
- Full-page scroll capture mode

## Phase 2 — Output modes & file size
- Keep current "full" mode (large, high-fidelity files)
- New **compact** default mode
  - Target 1024x800 viewport
  - Downscale / re-encode embedded images (e.g. JPEG quality reduction, resize to rendered dimensions)
  - Strip redundant precision from coordinates (round to 1 decimal)
  - Collapse empty/invisible elements
- CLI flag: `-mode full|compact`

## Phase 3 — Fidelity improvements
- Gradients (linear-gradient, radial-gradient backgrounds)
- Pseudo-elements (::before, ::after)
- SVG/canvas element passthrough
- Web font embedding (subset and inline)
- CSS transforms (translate, rotate, scale)
- Form controls (inputs, buttons, selects as visual approximations)

## Future ideas
- Streaming / incremental output for large pages
- Diff mode: capture two URLs and highlight visual differences
- PDF output via SVG → PDF conversion
- WASM build for in-browser use (no server-side Chrome needed)