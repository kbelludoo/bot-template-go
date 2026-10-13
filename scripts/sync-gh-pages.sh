#!/usr/bin/env bash
# Script para sincronizar telemetria com GitHub Pages a cada X minutos (via cron)
# MODO STEALTH: nicks, e-mails e IPs são anonimizados com jq ANTES do push.
set -e

PORT=${1:-3001}
DOCS_DIR="$(dirname "$0")/../docs"

mkdir -p "$DOCS_DIR"

STATUS=$(curl -s "http://127.0.0.1:${PORT}/api/status")
if [ -z "$STATUS" ] || [ "$STATUS" = "null" ]; then
  exit 0
fi

echo "$STATUS" | jq '
  (if (.accounts | type) == "array" then
    .accounts |= (to_entries | map(
        .key as $i
        | .value
        | (if has("name") then .name = ("Conta #" + (($i + 1) | tostring)) else . end)
        | (if has("trainerName") then .trainerName = ("Treinador #" + (($i + 1) | tostring)) else . end)
        | (if has("alias") then .alias = ("Treinador #" + (($i + 1) | tostring)) else . end)
        | (if has("charName") then .charName = ("Personagem #" + (($i + 1) | tostring)) else . end)
        | del(.bindIp, .outboundIp)
      ))
   else . end)
  | walk(
      if type == "string" and (test("^https?://") | not) then
        gsub("[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"; "usuario@privado.com")
        | gsub("\\b[0-9]{1,3}(\\.[0-9]{1,3}){3}\\b"; "x.x.x.x")
        | gsub("(?i)kbelludoo[0-9]*|lohmoreira|lrn"; "treinador")
      else . end
    )
' > "$DOCS_DIR/status.json.tmp"

mv "$DOCS_DIR/status.json.tmp" "$DOCS_DIR/status.json"

if [ -d ".git" ]; then
    git add "$DOCS_DIR/status.json"
    git commit -m "chore: sync telemetry $(date +'%Y-%m-%d %H:%M:%S')" || true
    git push origin main || true
fi
echo "Telemetria anonimizada sincronizada com o GitHub Pages!"
