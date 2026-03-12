#!/usr/bin/env bash
set -euo pipefail

MODULE="github.com/v4n5haj/discord-mass-dm-go"
VER="v0.0.0-20221119004734-95e8559721c4"
OUT="yikes-${VER}.zip"

# Known Go module mirrors (some may return 404 for this package)
proxies=(
  "https://proxy.golang.org"
  "https://goproxy.cn"
  "https://goproxy.io"
  "https://mirrors.aliyun.com/goproxy"
)

for p in "${proxies[@]}"; do
  echo "== Trying $p =="
  URL="$p/${MODULE}/@v/${VER}.zip"
  if curl -fsI "$URL" >/dev/null; then
    echo "Found at $p — downloading..."
    curl -fL "$URL" -o "$OUT"
    echo "Saved -> $OUT"
    exit 0
  else
    echo "Not available at $p"
  fi
done

echo "No mirror served ${MODULE}@${VER}. It may have been purged."
exit 2
