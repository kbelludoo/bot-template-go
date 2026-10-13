#!/usr/bin/env bash
bash /home/ubuntu/tibidle-bot/sync-gh-pages.sh
sleep 4
URL="https://kbelludoo.github.io/tibidle-bot/status.json?v=$(date +%s)"
curl -s "$URL" -o /tmp/t.json
echo "== campos com kbelludoo =="
jq -r 'paths(scalars) as $p | select(getpath($p) | tostring | test("kbelludoo"; "i")) | "\($p | join(".")) = \(getpath($p))"' /tmp/t.json
echo "== fim =="
