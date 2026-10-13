#!/usr/bin/env bash
# Sincroniza a telemetria do Sigma Bot para o GitHub Pages (docs/status.json)
# MODO ANONIMO: mascarar trainerName/alias/id e scrub de e-mails, IPs e nicks nos logs
STATUS=$(curl -s http://127.0.0.1:3002/api/status)
if [ -z "$STATUS" ] || [ "$STATUS" = "null" ]; then
  exit 0
fi

STATUS=$(echo "$STATUS" | jq '
  (if (.accounts | type) == "array" then
    .accounts |= (to_entries | map(
        .key as $i
        | .value
        | .id = ("acc" + (($i + 1) | tostring))
        | .trainerName = ("Treinador #" + (($i + 1) | tostring))
        | .alias = ("Treinador #" + (($i + 1) | tostring))
        | del(.bindIp, .outboundIp)
      ))
   else . end)
  | walk(
      if type == "string" and (test("^https?://") | not) then
        gsub("[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}"; "usuario@privado.com")
        | gsub("\\b[0-9]{1,3}(\\.[0-9]{1,3}){3}\\b"; "x.x.x.x")
        | gsub("(?i)kbelludoo[0-9]*|lohmoreira"; "treinador")
        | gsub("LrN"; "treinador")
      else . end
    )
')

if [ -z "$STATUS" ] || [ "$STATUS" = "null" ]; then
  exit 0
fi

B64=$(echo "$STATUS" | base64 -w 0)
SHA=$(gh api repos/kbelludoo/sigma-bot/contents/docs/status.json --jq .sha 2>/dev/null)

if [ -n "$SHA" ]; then
  gh api -X PUT repos/kbelludoo/sigma-bot/contents/docs/status.json \
    -f message="telemetry: live bot status update (stealth)" \
    -f content="$B64" \
    -f sha="$SHA" > /dev/null 2>&1
else
  gh api -X PUT repos/kbelludoo/sigma-bot/contents/docs/status.json \
    -f message="telemetry: live bot status update (stealth)" \
    -f content="$B64" > /dev/null 2>&1
fi
echo "Telemetria anonimizada sincronizada com o GitHub Pages!"
