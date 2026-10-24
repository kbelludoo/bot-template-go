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

  <script>
    function toggleCardFace(key) {
      var card = document.getElementById('card-' + key);
      if (card) {
        card.classList.toggle('is-flipped');
      }
    }

    var autoFlipCycle = 0;
    setInterval(function() {
      autoFlipCycle = 1 - autoFlipCycle;
      var cards = document.querySelectorAll('.bot-card:not(:hover)');
      cards.forEach(function(card) {
        if (autoFlipCycle === 1) {
          card.classList.add('is-flipped');
        } else {
          card.classList.remove('is-flipped');
        }
      });
    }, 8500);

    function formatNumber(n) {
      if (n < 1000) return n.toString();
      if (n < 1000000) return (n / 1000).toFixed(1) + 'k';
      return (n / 1000000).toFixed(2) + 'kk';
    }

    function formatRate(n) {
      if (!n || n <= 0) return 'Calculando...';
      return '+' + formatNumber(Math.round(n)) + '/h';
    }

    function formatUptime(sec, fallback) {
      if (sec && sec > 0) {
        var h = Math.floor(sec / 3600);
        var m = Math.floor((sec % 3600) / 60);
        return h + 'h ' + (m < 10 ? '0' : '') + m + 'm';
      }
      return fallback || '24/7 Ativo';
    }

    function getUnifiedAccounts(b) {
      if (!b || !b.status) return [];

      // Kick Miner
      if (b.status.accounts && typeof b.status.accounts === 'object') {
        var accKeys = Object.keys(b.status.accounts);
        if (accKeys.length > 0 && b.status.accounts[accKeys[0]] && b.status.accounts[accKeys[0]].streamers) {
          return accKeys.sort().map(function(k) {
            var entry = b.status.accounts[k];
            var streams = entry.streamers || {};
            var sKeys = Object.keys(streams).sort();
            var totalPts = 0;
            var liveCount = 0;
            var liveNames = [];
            for (var si = 0; si < sKeys.length; si++) {
              var st = streams[sKeys[si]];
              totalPts += (st.points || 0);
              if (st.isLive) {
                liveCount++;
                if (liveNames.length < 3) liveNames.push(st.streamer || sKeys[si]);
              }
            }
            var areaText = (entry.activeStreams || liveCount) + ' canais ativos';
            if (liveNames.length > 0) areaText = 'Ao vivo: ' + liveNames.join(', ') + ' (+' + liveCount + ')';
            return {
              name: 'Conta Kick #1',
              level: 0,
              vocation: 'Kick Stream',
              currentArea: areaText,
              gold: totalPts,
              kills: liveCount,
              status: 'Minerando',
              isStream: true,
              durationSec: 0
            };
          });
        }
      }

      if (Array.isArray(b.status.accounts) && b.status.accounts.length > 0) {
        return b.status.accounts;
      }
      if (b.status.accounts && typeof b.status.accounts === 'object' && Object.keys(b.status.accounts).length > 0) {
        return Object.values(b.status.accounts);
      }
      if (b.status.bots && typeof b.status.bots === 'object') {
        var keys = Object.keys(b.status.bots).sort();
        return keys.map(function(k) {
          var entry = b.status.bots[k];
          var s = entry.status || {};
          var tel = s.telemetry || {};
          var rates = s.rates || {};
          var statusStr = s.status || 'online';
          if (s.treino) statusStr = 'Treino';
          return {
            name: entry.label || entry.name || k,
            level: s.level || tel.level || 1,
            vocation: s.vocation || '',
            currentArea: s.currentArea || (s.hunt && s.hunt.name) || (s.treino ? 'Treino de Skills' : ''),
            gold: s.gold || tel.gold || 0,
            kills: s.kills || tel.kills || 0,
            stamina: tel.stamina || '',
            status: statusStr,
            durationSec: s.uptimeSec || 0,
            expPerHour: rates.xpPerHour || 0,
            goldPerHour: rates.goldPerHour || 0
          };
        });
      }
      if (b.status.statuses && typeof b.status.statuses === 'object') {
        var keys = Object.keys(b.status.statuses).sort();
        return keys.map(function(k) {
          var entry = b.status.statuses[k];
          var streams = entry.streamers || {};
          var watching = Object.values(streams).filter(function(x) { return x.watching; }).map(function(x) { return x.displayName || x.login; });
          var liveCount = Object.values(streams).filter(function(x) { return x.isLive; }).length;
          var totalPts = 0;
          Object.values(streams).forEach(function(x) { totalPts += (x.points || 0); });
          return {
            name: entry.username || k,
            level: 0,
            vocation: 'Twitch Drops',
            currentArea: watching.length ? 'Assistindo ' + watching.join(', ') : 'Minerando canais Twitch',
            gold: totalPts,
            kills: liveCount,
            status: 'Online',
            isStream: true,
            durationSec: 0
          };
        });
      }

      if (b.status.uptime || b.status.ramSysMB || b.status.goroutines) {
        return [{
          name: 'Sessão Conectada',
          level: 1,
          vocation: 'Operação Contínua',
          currentArea: 'Monitorando em Standby',
          status: 'Standby',
          durationSec: 0
        }];
      }

      return [];
    }

    async function refreshHub() {
      try {
        var res = await fetch('status.json?t=' + Date.now());
        if (!res.ok) return;
        var data = await res.json();
        if (!data || !data.bots) return;

        if (data.vps) {
          var v = data.vps;
          var ramVal = document.getElementById('ramVal');
          var ramProg = document.getElementById('ramProgress');
          var ramSub = document.getElementById('ramSub');
          if (ramVal) ramVal.textContent = (v.ramUsedMb || 0) + ' / ' + (v.ramTotalMb || 0) + ' MB';
          if (ramProg) ramProg.style.width = (v.ramPercent || 0) + '%';
          if (ramSub) ramSub.textContent = 'Livre: ' + (v.ramAvailableMb || 0) + ' MB (' + (v.ramPercent || 0) + '% em uso)';
        }

        var botKeys = Object.keys(data.bots);
        var activeCount = botKeys.filter(function(k) { return getUnifiedAccounts(data.bots[k]).length > 0; }).length;
        var botsCountEl = document.getElementById('botsCountVal');
        if (botsCountEl) botsCountEl.textContent = activeCount + ' / ' + botKeys.length;

        for (var i = 0; i < botKeys.length; i++) {
          var k = botKeys[i];
          var b = data.bots[k];
          var accList = document.getElementById('acclist-' + k);
          var badge = document.getElementById('badge-' + k);
          var accounts = getUnifiedAccounts(b);
          var isStream = (k === 'kickminer' || k === 'twitchminer');
          var isService = (k === 'pocketweb' || k === 'kaggle' || k === 'telegram');

          if (badge && accounts.length > 0) {
            if (isStream) {
              var wave = '<span class="live-wave"><span></span><span></span><span></span></span>';
              badge.innerHTML = wave + (accounts[0].kills > 0 ? accounts[0].kills + ' AO VIVO' : 'MINERANDO');
            } else if (isService) {
              badge.textContent = accounts[0].status ? accounts[0].status.toUpperCase() : 'ONLINE';
            } else {
              badge.textContent = accounts.length > 1 ? accounts.length + '/' + accounts.length + ' ATIVAS' : (accounts[0].status ? accounts[0].status.toUpperCase() : 'ONLINE');
            }
          }

          if (accList && accounts.length > 0) {
            var botRam = (b.status && b.status.ramSysMB) ? b.status.ramSysMB.toFixed(1) + ' MB' : '< 10 MB';
            var botUptime = (b.status && b.status.uptime) ? b.status.uptime : '24/7 Ativo';

            var frontHtml = accounts.map(function(a) {
              var s = '<div class="acc-item">';
              if (isStream) {
                s += '<div class="acc-item-header"><span>' + (a.name || 'Conta') + '</span><span style="font-size: 0.72rem; font-weight: 700;">STREAM VIEWER</span></div>';
                if (a.currentArea) {
                  s += '<div class="stat-row" style="padding: 4px 8px; border-radius: 6px; margin-bottom: 4px; background: rgba(255,255,255,0.05); font-weight: 700;">📡 ' + a.currentArea + '</div>';
                }
                s += '<div class="stat-row"><span style="color: #facc15; font-weight: 800;">⭐ ' + (a.gold || 0).toLocaleString() + ' pontos</span><span style="color: #38bdf8; font-weight: 700;">🔴 ' + (a.kills || 0) + ' ' + ((a.kills === 1) ? 'canal ao vivo' : 'canais ao vivo') + '</span></div>';
              } else if (isService) {
                s += '<div class="acc-item-header"><span>' + (a.name || 'Serviço') + '</span><span style="font-size: 0.72rem; font-weight: 700;">' + (a.vocation || 'Service') + '</span></div>';
                if (a.currentArea) {
                  s += '<div class="stat-row" style="padding: 4px 8px; border-radius: 6px; margin-bottom: 4px; background: rgba(255,255,255,0.05); font-weight: 700;">⚙️ ' + a.currentArea + '</div>';
                }
                s += '<div class="stat-row"><span style="color: #4ade80; font-weight: 700;">🟢 ' + (a.status || 'Online') + '</span><span style="color: #38bdf8; font-weight: 600;">⚡ Operação Contínua</span></div>';
              } else {
                s += '<div class="acc-item-header"><span>' + (a.name || 'Conta') + ' (Nv. ' + (a.level || 1) + ')</span><span>' + (a.vocation || '') + '</span></div>';
                if (a.currentArea) {
                  s += '<div class="stat-row" style="padding: 4px 8px; border-radius: 6px; margin-bottom: 4px; background: rgba(255,255,255,0.05); font-weight: 700;">🗺️ ' + a.currentArea + '</div>';
                }
                if (a.maxHp) {
                  s += '<div class="stat-row"><span>❤️ HP</span><span style="color: #4ade80;">' + (a.hp || 0).toLocaleString() + ' / ' + (a.maxHp || 0).toLocaleString() + '</span></div>';
                }
                if (a.stamina) {
                  var stam = typeof a.stamina === 'number' ? a.stamina + '%' : a.stamina;
                  s += '<div class="stat-row"><span>⚡ Stamina</span><span style="color: #38bdf8; font-weight: 600;">' + stam + '</span></div>';
                }
                s += '<div class="stat-row"><span style="color: #facc15; font-weight: 800;">💰 ' + (a.gold || 0).toLocaleString() + ' gp</span><span style="color: #38bdf8;">⚔️ ' + (a.kills || 0).toLocaleString() + ' kills</span></div>';
              }
              s += '</div>';
              return s;
            }).join('');

            var backHtml = accounts.map(function(a) {
              var s = '<div class="acc-item">';
              if (isStream) {
                s += '<div class="acc-item-header"><span>⚡ Telemetria Drops</span><span class="perf-tag">AUTOMÁTICO</span></div>';
                s += '<div class="stat-row"><span>🎁 Coleta de Pontos</span><span style="color: #4ade80; font-weight: 700;">Ativo (Zero-Video WS)</span></div>';
                s += '<div class="stat-row"><span>🧠 Memória Bot (Go)</span><span style="color: #38bdf8; font-weight: 600;">' + botRam + '</span></div>';
                s += '<div class="stat-row"><span>⏱️ Uptime Contínuo</span><span style="color: #facc15;">' + formatUptime(a.durationSec, botUptime) + '</span></div>';
                s += '<div class="stat-row"><span>🛡️ Camuflagem IRC</span><span style="color: #a855f7; font-weight: 700;">100% Stealth Mode</span></div>';
              } else if (isService) {
                s += '<div class="acc-item-header"><span>⚡ Telemetria de Sistema</span><span class="perf-tag">ONLINE 24/7</span></div>';
                s += '<div class="stat-row"><span>🔄 Healthcheck</span><span style="color: #4ade80; font-weight: 700;">🟢 Heartbeat OK</span></div>';
                s += '<div class="stat-row"><span>🧠 Memória Isolada</span><span style="color: #38bdf8; font-weight: 600;">' + botRam + '</span></div>';
                s += '<div class="stat-row"><span>🛡️ Supervisor Resiliente</span><span style="color: #a855f7; font-weight: 700;">Goroutines com Recover</span></div>';
                s += '<div class="stat-row"><span>⏱️ Uptime</span><span style="color: #facc15;">' + formatUptime(a.durationSec, botUptime) + '</span></div>';
              } else {
                s += '<div class="acc-item-header"><span>⚡ ' + (a.name || 'Conta') + ' — Rendimento</span><span class="perf-tag">PROTEGIDO</span></div>';
                s += '<div class="stat-row"><span>📈 EXP / Hora</span><span style="color: #4ade80; font-weight: 700;">' + formatRate(a.expPerHour) + '</span></div>';
                s += '<div class="stat-row"><span>💰 Ouro / Hora</span><span style="color: #facc15; font-weight: 700;">' + formatRate(a.goldPerHour) + '</span></div>';
                s += '<div class="stat-row"><span>⏱️ Tempo de Caça</span><span style="color: #38bdf8;">' + formatUptime(a.durationSec, botUptime) + '</span></div>';
                s += '<div class="stat-row"><span>🧠 Memória Bot (Go)</span><span style="color: #94a3b8;">' + botRam + '</span></div>';
              }
              s += '</div>';
              return s;
            }).join('');

            accList.innerHTML = '<div class="card-face face-gameplay">' + frontHtml + '</div>' +
                                '<div class="card-face face-performance">' + backHtml + '</div>';
          }
        }
      } catch (e) {
        console.warn('Erro ao atualizar hub:', e);
      }
    }
    setInterval(refreshHub, 12000);
  </script>
</body>
</html>`
