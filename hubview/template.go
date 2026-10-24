package hubview

const hubTemplate = `<!DOCTYPE html>
<html lang="pt-BR" class="notranslate" translate="no">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>⚡ Bot Hub — VPS Command Center & Game/Stream/Service Hub (Go Engine)</title>
  <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>⚡</text></svg>">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap" rel="stylesheet">
  <style>` + hubStyles + `</style>
</head>
<body>
  <div class="container">
    <header class="header">
      <div class="header-left">
        <div class="logo-badge">⚡</div>
        <div class="header-title">
          <h1>VPS Command Center</h1>
          <p><span>🌐 Cluster OCI Always-Free</span> • <span>🚀 Go 1.27 Native Engine</span></p>
        </div>
      </div>
      <div class="header-right">
        <button class="test-levelup-badge" onclick="testLevelUp()" title="Clique para testar a celebração Fullscreen de Level Up (4 segundos)">
          🎉 TESTAR LEVEL UP
        </button>
        <div class="live-badge">
          <span class="pulse-dot"></span>
          <span>LIVE TELEMETRY</span>
        </div>
      </div>
    </header>

    <div class="vps-metrics-grid">
      <div class="vps-card">
        <div class="vps-label">Serviços em Operação</div>
        <div class="vps-val" id="botsCountVal" style="color: var(--accent);">{{.ActiveBots}} / {{.TotalBots}}</div>
        <div class="vps-sub">Bots & Miners conectados 24/7 na VPS</div>
      </div>
      <div class="vps-card">
        <div class="vps-label">Memória RAM VPS</div>
        <div class="vps-val" id="ramVal" style="color: var(--green);">{{.VPS.RAMUsedMB}} / {{.VPS.RAMTotalMB}} MB</div>
        <div class="progress-bar"><div class="progress-fill fill-ram" id="ramProgress" style="width: {{.VPS.RAMPercent}}%;"></div></div>
        <div class="vps-sub" id="ramSub">Livre: {{.VPS.RAMAvailableMB}} MB ({{.VPS.RAMPercent}}% em uso)</div>
      </div>
      <div class="vps-card">
        <div class="vps-label">Uptime & Conexão</div>
        <div class="vps-val" style="color: var(--gold); font-size: 1.15rem;">OCI Ashburn</div>
        <div class="vps-sub" id="lastUpdatedSub">{{.TimeAgo}}</div>
      </div>
    </div>

    <div class="bots-grid" id="botsGrid">
      {{range .Bots}}
      {{$bot := .}}
      <div class="bot-card" id="card-{{$bot.Meta.Key}}" style="--theme: {{css $bot.Meta.Theme}}; --theme-bg: {{css $bot.Meta.ThemeBg}}; --theme-border: {{css $bot.Meta.ThemeBorder}}; --theme-glow: {{css $bot.Meta.ThemeGlow}};">
        <div>
          <div class="bot-header">
            <div class="bot-name-group">
              <h3>{{$bot.Meta.Icon}} {{$bot.Meta.Name}}</h3>
              <span class="bot-port">
                {{if gt $bot.Meta.Port 0}}PORTA {{$bot.Meta.Port}} • {{end}}
                {{if $bot.Meta.IsStream}}STREAM HARVESTER{{else if $bot.Meta.IsService}}BACKGROUND AGENT{{else}}GO NATIVE{{end}}
              </span>
            </div>
            <div class="face-header-controls">
              <button class="face-toggle-btn" onclick="toggleCardFace('{{$bot.Meta.Key}}')" title="Clique para alternar visão Gameplay / Rendimento">
                🔄 <span id="facetext-{{$bot.Meta.Key}}">VISÃO</span>
              </button>
              <span class="badge-status" id="badge-{{$bot.Meta.Key}}">
                {{if $bot.Meta.IsStream}}
                <span class="live-wave"><span></span><span></span><span></span></span>
                {{end}}
                {{$bot.BadgeText}}
              </span>
            </div>
          </div>

          <div class="acc-list" id="acclist-{{$bot.Meta.Key}}">
            {{if $bot.Accounts}}
              <!-- FACE 1: GAMEPLAY / COMBATE -->
              <div class="card-face face-gameplay">
                {{range $idx, $acc := $bot.Accounts}}
                <div class="acc-item">
                  {{if $bot.Meta.IsStream}}
                  <div class="acc-item-header">
                    <span>{{$bot.Meta.Icon}} {{$acc.Name}}</span>
                    <span style="color: {{css $bot.Meta.Theme}}; font-size: 0.72rem; font-weight: 700;">STREAM VIEWER</span>
                  </div>
                  {{if $acc.CurrentArea}}
                  <div class="stat-row" style="background: {{css $bot.Meta.ThemeBg}}; padding: 4px 8px; border-radius: 6px; margin-bottom: 4px;">
                    <span style="font-weight: 700; font-size: 0.72rem;">📡 {{$acc.CurrentArea}}</span>
                  </div>
                  {{end}}
                  <div class="stat-row">
                    <span style="color: #facc15; font-weight: 800;">⭐ {{formatNumber $acc.Gold}} pontos</span>
                    <span style="color: #38bdf8; font-weight: 700;">🔴 {{$acc.Kills}} {{if eq $acc.Kills 1}}canal ao vivo{{else}}canais ao vivo{{end}}</span>
                  </div>
                  {{else if $bot.Meta.IsService}}
                  <div class="acc-item-header">
                    <span>{{$bot.Meta.Icon}} {{$acc.Name}}</span>
                    <span style="color: {{css $bot.Meta.Theme}}; font-size: 0.72rem; font-weight: 700;">{{$acc.Vocation}}</span>
                  </div>
                  {{if $acc.CurrentArea}}
                  <div class="stat-row" style="background: {{css $bot.Meta.ThemeBg}}; padding: 4px 8px; border-radius: 6px; margin-bottom: 4px;">
                    <span style="font-weight: 700; font-size: 0.72rem;">⚙️ {{$acc.CurrentArea}}</span>
                  </div>
                  {{end}}
                  <div class="stat-row">
                    <span style="color: #4ade80; font-weight: 700;">🟢 {{$acc.Status}}</span>
                    <span style="color: #38bdf8; font-weight: 600;">⚡ Operação Contínua</span>
                  </div>
                  {{else}}
                  <div class="acc-item-header">
                    <span>{{$bot.Meta.Icon}} {{$acc.Name}} (Nv. {{$acc.Level}})</span>
                    <span style="color: {{css $bot.Meta.Theme}};">{{$acc.Vocation}}</span>
                  </div>
                  {{if $acc.CurrentArea}}
                  <div class="stat-row" style="background: {{css $bot.Meta.ThemeBg}}; padding: 4px 8px; border-radius: 6px; margin-bottom: 4px;">
                    <span style="font-weight: 700; font-size: 0.72rem;">🗺️ {{$acc.CurrentArea}}</span>
                  </div>
                  {{end}}
                  {{if gt $acc.MaxHP 0}}
                  <div class="stat-row">
                    <span>❤️ HP</span>
                    <span style="color: #4ade80;">{{$acc.HP}} / {{$acc.MaxHP}}</span>
                  </div>
                  {{end}}
                  {{if $acc.Stamina}}
                  <div class="stat-row">
                    <span>⚡ Stamina</span>
                    <span style="color: #38bdf8; font-weight: 600;">{{$acc.Stamina}}</span>
                  </div>
                  {{end}}
                  <div class="stat-row">
                    <span style="color: #facc15; font-weight: 800;">💰 {{formatNumber $acc.Gold}} gp</span>
                    <span style="color: #38bdf8; font-weight: 700;">⚔️ {{$acc.Kills}} kills</span>
                  </div>
                  {{end}}
                </div>
                {{end}}
              </div>

              <!-- FACE 2: RENDIMENTO & PERFORMANCE -->
              <div class="card-face face-performance">
                {{range $idx, $acc := $bot.Accounts}}
                <div class="acc-item">
                  {{if $bot.Meta.IsStream}}
                  <div class="acc-item-header">
                    <span>⚡ Telemetria Drops</span>
                    <span class="perf-tag">AUTOMÁTICO</span>
                  </div>
                  <div class="stat-row">
                    <span>🎁 Coleta de Pontos</span>
                    <span style="color: #4ade80; font-weight: 700;">Ativo (Zero-Video WS)</span>
                  </div>
                  <div class="stat-row">
                    <span>🧠 Memória Bot (Go)</span>
                    <span style="color: #38bdf8; font-weight: 600;">{{if gt $bot.Bot.Status.RAMSysMB 0.0}}{{printf "%.1f MB" $bot.Bot.Status.RAMSysMB}}{{else if gt $bot.Bot.Status.RAMMB 0.0}}{{printf "%.1f MB" $bot.Bot.Status.RAMMB}}{{else}}&lt; 15 MB{{end}}</span>
                  </div>
                  <div class="stat-row">
                    <span>⏱️ Uptime Contínuo</span>
                    <span style="color: #facc15;">{{formatUptime $acc.DurationSec $bot.Bot.Status.Uptime}}</span>
                  </div>
                  <div class="stat-row">
                    <span>🛡️ Camuflagem IRC</span>
                    <span style="color: #a855f7; font-weight: 700;">100% Stealth Mode</span>
                  </div>
                  {{else if $bot.Meta.IsService}}
                  <div class="acc-item-header">
                    <span>⚡ Telemetria de Sistema</span>
                    <span class="perf-tag">ONLINE 24/7</span>
                  </div>
                  <div class="stat-row">
                    <span>🔄 Healthcheck</span>
                    <span style="color: #4ade80; font-weight: 700;">🟢 Heartbeat OK</span>
                  </div>
                  <div class="stat-row">
                    <span>🧠 Memória Isolada</span>
                    <span style="color: #38bdf8; font-weight: 600;">{{if gt $bot.Bot.Status.RAMSysMB 0.0}}{{printf "%.1f MB" $bot.Bot.Status.RAMSysMB}}{{else}}Baixo Consumo{{end}}</span>
                  </div>
                  <div class="stat-row">
                    <span>🛡️ Supervisor Resiliente</span>
                    <span style="color: #a855f7; font-weight: 700;">Goroutines com Recover</span>
                  </div>
                  <div class="stat-row">
                    <span>⏱️ Uptime</span>
                    <span style="color: #facc15;">{{formatUptime $acc.DurationSec $bot.Bot.Status.Uptime}}</span>
                  </div>
                  {{else}}
                  <div class="acc-item-header">
                    <span>⚡ {{if $acc.Name}}{{$acc.Name}}{{else}}Conta{{end}} — Rendimento</span>
                    <span class="perf-tag">PROTEGIDO</span>
                  </div>
                  <div class="stat-row">
                    <span>📈 EXP / Hora</span>
                    <span style="color: #4ade80; font-weight: 700;">{{formatRate $acc.ExpPerHour}}</span>
                  </div>
                  <div class="stat-row">
                    <span>💰 Ouro / Hora</span>
                    <span style="color: #facc15; font-weight: 700;">{{formatRate $acc.GoldPerHour}}</span>
                  </div>
                  <div class="stat-row">
                    <span>⏱️ Tempo de Caça</span>
                    <span style="color: #38bdf8;">{{formatUptime $acc.DurationSec $bot.Bot.Status.Uptime}}</span>
                  </div>
                  <div class="stat-row">
                    <span>🧠 Memória Bot (Go)</span>
                    <span style="color: #94a3b8;">{{if gt $bot.Bot.Status.RAMSysMB 0.0}}{{printf "%.1f MB" $bot.Bot.Status.RAMSysMB}}{{else if gt $bot.Bot.Status.RAMMB 0.0}}{{printf "%.1f MB" $bot.Bot.Status.RAMMB}}{{else}}&lt; 10 MB{{end}}</span>
                  </div>
                  {{end}}
                </div>
                {{end}}
              </div>
            {{else}}
              <div style="color: var(--text-dim); font-size: 0.8rem;">Carregando telemetria...</div>
            {{end}}
          </div>
        </div>

        <div class="bot-footer">
          <span style="font-size: 0.72rem; color: var(--text-dim);" id="summary-{{$bot.Meta.Key}}">{{$bot.SummaryText}}</span>
          <a href="{{$bot.Meta.DashboardURL}}" target="_blank" class="btn-link">Abrir Dashboard ↗</a>
        </div>
      </div>
      {{end}}
    </div>

    <div class="stream-ticker-wrap">
      <div class="ticker-badge">🔴 LIVE FLEET</div>
      <div class="stream-ticker-track">
        <span>⚡ VPS CLUSTER ONLINE • 18 SERVIÇOS & BOTS NATIVOS EM GO • ZERO NODE.JS</span>
        <span>🎮 JOGOS, 📺 STREAMS & ⚙️ AGENTES: {{.TickerBotsList}}</span>
        <span>⚡ VPS CLUSTER ONLINE • 18 SERVIÇOS & BOTS NATIVOS EM GO • ZERO NODE.JS</span>
        <span>🎮 JOGOS, 📺 STREAMS & ⚙️ AGENTES: {{.TickerBotsList}}</span>
      </div>
    </div>
  </div>

  <!-- Fullscreen Level Up Spotlight Modal -->
  <div class="levelup-overlay" id="levelUpOverlay" onclick="dismissLevelUp()">
    <div class="levelup-modal" onclick="event.stopPropagation()">
      <button class="levelup-close-btn" onclick="dismissLevelUp()" title="Fechar">✕</button>
      <div class="levelup-tag">⚡ LEVEL UP! ⚡</div>
      
      <!-- Character Celebration GIF -->
      <div class="levelup-char-gif-wrap">
        <img id="lvlCharGif" class="levelup-char-gif" src="" alt="Personagem" onerror="this.style.display='none'; document.getElementById('lvlFallbackEmoji').style.display='block';">
        <div id="lvlFallbackEmoji" style="display:none; font-size: 3rem;">🏆</div>
      </div>

      <div class="levelup-title">NOVO NÍVEL ALCANÇADO</div>
      <div class="levelup-game-name" id="lvlGameName">⚔️ Jogo</div>
      <div class="levelup-level-row">
        <div class="levelup-level-box">
          <div class="levelup-level-lbl">Nível Anterior</div>
          <div class="levelup-level-num" id="lvlOldVal">Nv. 0</div>
        </div>
        <div class="levelup-arrow">➜</div>
        <div class="levelup-level-box new-level">
          <div class="levelup-level-lbl">Novo Nível</div>
          <div class="levelup-level-num" id="lvlNewVal">Nv. 0</div>
        </div>
      </div>
      <div class="levelup-info-box">
        <span id="lvlCharInfo">⚔️ Personagem</span>
        <span id="lvlAreaInfo">🗺️ Local</span>
      </div>
      <div class="levelup-timer-bar">
        <div class="levelup-timer-fill" id="lvlTimerFill"></div>
      </div>
    </div>
  </div>

  <script>` + hubScripts + `</script>
</body>
</html>`
