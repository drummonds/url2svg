# Changelog

## [Unreleased]

## [0.1.8] - 2026-09-25

 - Fix self update downgrading when the proxy lags

### Fixed
- `self update` never downgrades: it compares versions semantically, so a module proxy that has not yet seen a fresh tag leaves the newer installed binary alone.

## [0.1.7] - 2026-09-25

 - Version flag, self update, sanitised links and inline SVG, docs favicon

### Added
- `-version` flag and `url2svg self update`, which reinstalls the latest release from the Go module proxy.
- Docs favicon, captured by url2svg itself from a badge page served by the demo server.

## [0.1.6] - 2026-09-25

 - Docs on docs.bytestone.uk, module moved to git.bytestone.uk

### Security
- Links keep only `http`, `https` and `mailto` schemes; `javascript:` and other executable hrefs are dropped.
- Inline `<svg>` elements are scrubbed of `<script>`, `on*` handlers and executable links before being embedded.

### Changed
- Module path is now `git.bytestone.uk/hum3/url2svg`; the Codeberg repo is retired and GitHub is a mirror.
- Documentation moved to https://url2svg.docs.bytestone.uk/ and shows the demo capture scaled down to icon size.
- `task docs:build` no longer hangs after capturing the demo page.

### Fixed
- Inline `<svg>` elements are captured as a single image at their rendered size. Previously every shape became its own `<image>` and each `<text>` was duplicated as an HTML text run at the unscaled CSS font size, so chart labels rendered oversized.

## [0.1.5] - 2026-04-01

 - Enable install verification on tp release

## [0.1.4] - 2026-03-31

## [0.1.3] - 2026-03-31

 - Update CHANGELOG for v0.1.2

## [0.1.2] - 2026-03-30

 - adding gitgnore
