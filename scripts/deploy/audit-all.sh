#!/usr/bin/env bash
node /home/ubuntu/sync-individual-stealth.js
sleep 5
for repo in sigma-bot tibidle-bot leveling-bot; do
  echo "-- $repo --"
  url="https://kbelludoo.github.io/$repo/status.json?v=$(date +%s)"
  curl -s "$url" | jq -r '.. | strings' | grep -i -E 'kbelludoo|lohmoreira|@gmail|@hotmail|@outlook' || echo "LIMPO"
done
echo "== ids tibidle =="
curl -s "https://kbelludoo.github.io/tibidle-bot/status.json?v=$(date +%s)" | jq -c '[.accounts[]? | {name, id}]'
