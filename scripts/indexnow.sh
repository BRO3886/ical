#!/usr/bin/env sh
# Submit this site's live URLs to IndexNow (Bing/Microsoft + Yandex/Naver/Seznam/Yep).
set -eu
HOST="ical.sidv.dev"
KEY="9243cdd67ce6db495ec7acddec1dec27"
KEYLOC="https://${HOST}/${KEY}.txt"
ENDPOINT="https://api.indexnow.org/indexnow"
urls=$(curl -fsS "https://${HOST}/sitemap.xml" | grep -o '<loc>[^<]*</loc>' | sed 's/<loc>//; s|</loc>||')
[ -n "$urls" ] || { echo "no URLs found in sitemap"; exit 1; }
count=0; fail=0
for u in $urls; do
  code=$(curl -s -o /dev/null -w '%{http_code}' "${ENDPOINT}?url=${u}&key=${KEY}&keyLocation=${KEYLOC}")
  case "$code" in 200|202) count=$((count+1));; *) echo "FAILED ($code): $u"; fail=1;; esac
done
echo "Submitted ${count} URL(s) to IndexNow"
[ "$fail" -eq 0 ] || exit 1
