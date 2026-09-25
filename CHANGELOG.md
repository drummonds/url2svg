# Changelog

## [Unreleased]

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
