#!/usr/bin/env bash
# Generate the desktop app icon set from the Maestro logo.
#
# Source:  maestro-logo.svg  (repo root)
# Outputs: app/assets/icon.png (1024), icon.icns (macOS), icon.ico (Windows)
#
# Requirements: rsvg-convert, iconutil (macOS), and ImageMagick (magick/convert).
# Usage:  ./scripts/generate-icons.sh
set -euo pipefail

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
src="$repo_root/maestro-logo.svg"
assets="$repo_root/app/assets"
mkdir -p "$assets"

echo "Rendering 1024px master PNG…"
rsvg-convert -w 1024 -h 1024 "$src" -o "$assets/icon.png"

echo "Building macOS .icns…"
iconset="$(mktemp -d)/icon.iconset"
mkdir -p "$iconset"
for size in 16 32 128 256 512; do
  rsvg-convert -w "$size"  -h "$size"  "$src" -o "$iconset/icon_${size}x${size}.png"
  dbl=$((size * 2))
  rsvg-convert -w "$dbl" -h "$dbl" "$src" -o "$iconset/icon_${size}x${size}@2x.png"
done
iconutil -c icns "$iconset" -o "$assets/icon.icns"

echo "Building Windows .ico…"
magick_bin="$(command -v magick || command -v convert)"
"$magick_bin" -background none "$src" \
  -define icon:auto-resize=16,32,48,64,128,256 "$assets/icon.ico"

echo "Done. Wrote:"
ls -la "$assets/icon.png" "$assets/icon.icns" "$assets/icon.ico"
