#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:3000}"
KEYWORDS=("热播" "哪吒" "流浪地球" "庆余年" "海贼王" "甄嬛传")

if [[ "$#" -gt 0 ]]; then
  KEYWORDS=("$@")
fi

echo "[warm-image-cache] base=${BASE_URL}"

for kw in "${KEYWORDS[@]}"; do
  echo "[warm-image-cache] search keyword: ${kw}"
  payload="$(curl -fsSL "${BASE_URL}/api/search?q=$(python3 -c 'import urllib.parse,sys;print(urllib.parse.quote(sys.argv[1]))' "${kw}")")"

  python3 - "$BASE_URL" <<'PY' <<<"$payload"
import json
import sys
import urllib.parse
import urllib.request

base = sys.argv[1]
raw = sys.stdin.read().strip()
if not raw:
    sys.exit(0)

try:
    data = json.loads(raw)
except Exception:
    sys.exit(0)

results = data.get("results") or []
urls = []
for item in results[:80]:
    poster = (item or {}).get("poster")
    if isinstance(poster, str) and poster.strip().startswith(("http://", "https://")):
        urls.append(poster.strip())

seen = set()
for poster in urls:
    if poster in seen:
        continue
    seen.add(poster)
    target = f"{base}/api/image?url={urllib.parse.quote(poster, safe='')}"
    req = urllib.request.Request(target, method="GET")
    try:
        with urllib.request.urlopen(req, timeout=12) as resp:
            _ = resp.read(1)
            print(f"warmed {resp.status} {poster[:120]}")
    except Exception:
        pass
PY
done

echo "[warm-image-cache] done"
