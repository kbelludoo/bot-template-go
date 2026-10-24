package hubview

const hubTemplate = `<!DOCTYPE html>
<html lang="pt-BR" class="notranslate" translate="no">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>⚡ Bot Hub — VPS Command Center & Game/Stream Bot Kit (Go Engine)</title>
  <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>⚡</text></svg>">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #07090e;
      --card-bg: rgba(15, 23, 42, 0.75);
      --card-border: rgba(255, 255, 255, 0.08);
      --card-hover: rgba(255, 255, 255, 0.14);
      --accent: #00f0ff;
      --accent-glow: rgba(0, 240, 255, 0.25);
      --green: #10b981;
      --green-glow: rgba(16, 185, 129, 0.25);
      --gold: #f59e0b;
      --yellow: #f59e0b;
      --red: #ef4444;
      --purple: #a855f7;
      --cyan: #06b6d4;
      --text: #f8fafc;
      --text-muted: #94a3b8;
      --text-dim: #64748b;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: var(--bg);
      background-image: 
        radial-gradient(circle at 10% 20%, rgba(0, 240, 255, 0.05) 0%, transparent 40%),
        radial-gradient(circle at 90% 10%, rgba(168, 85, 247, 0.06) 0%, transparent 45%),
        radial-gradient(circle at 50% 80%, rgba(16, 185, 129, 0.04) 0%, transparent 50%),
        linear-gradient(to bottom, #07090e, #0a0e1a);
      color: var(--text);
      font-family: 'Plus Jakarta Sans', -apple-system, sans-serif;
      min-height: 100vh;
      padding: 24px;
      line-height: 1.5;
    }
    .container { max-width: 1360px; margin: 0 auto; }
    .header {
      display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 16px;
      padding: 22px 28px; background: var(--card-bg); backdrop-filter: blur(16px);
      border: 1px solid var(--card-border); border-radius: 20px; margin-bottom: 24px;
      box-shadow: 0 10px 35px -10px rgba(0, 0, 0, 0.6);
    }
    .header-left { display: flex; align-items: center; gap: 16px; }
    .logo-badge {
      width: 52px; height: 52px; border-radius: 14px;
      background: linear-gradient(135deg, rgba(0, 240, 255, 0.2), rgba(139, 92, 246, 0.3));
      border: 1px solid var(--accent); display: flex; align-items: center; justify-content: center;
      font-size: 1.7rem; box-shadow: 0 0 20px var(--accent-glow);
    }
    .header-title h1 {
      font-size: 1.45rem; font-weight: 800; letter-spacing: -0.02em;
      background: linear-gradient(90deg, #ffffff, #94a3b8);
      -webkit-background-clip: text; -webkit-text-fill-color: transparent;
    }
    .header-title p {
      font-size: 0.8rem; color: var(--text-muted); font-family: 'JetBrains Mono', monospace;
      display: flex; align-items: center; gap: 8px; margin-top: 2px;
    }
    .header-right { display: flex; align-items: center; gap: 12px; }
    .live-badge {
      display: inline-flex; align-items: center; gap: 8px; padding: 7px 16px; border-radius: 999px;
      font-size: 0.78rem; font-weight: 700; font-family: 'JetBrains Mono', monospace;
      background: rgba(16, 185, 129, 0.12); border: 1px solid var(--green); color: var(--green);
    }
    .pulse-dot {
      width: 8px; height: 8px; background: var(--green); border-radius: 50%;
      box-shadow: 0 0 8px var(--green);
    }
    .vps-metrics-grid {
      display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 16px; margin-bottom: 28px;
    }
    .vps-card {
      background: var(--card-bg); backdrop-filter: blur(12px); border: 1px solid var(--card-border);
      border-radius: 16px; padding: 18px 22px; transition: transform 0.2s ease;
    }
    .vps-card:hover { transform: translateY(-2px); border-color: var(--card-hover); }
    .vps-label {
      font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.08em;
      color: var(--text-dim); font-weight: 700; margin-bottom: 8px;
    }
    .vps-val { font-size: 1.5rem; font-weight: 800; font-family: 'JetBrains Mono', monospace; }
    .vps-sub { font-size: 0.75rem; color: var(--text-muted); margin-top: 6px; }
    .progress-bar {
      width: 100%; height: 6px; background: rgba(255, 255, 255, 0.08);
      border-radius: 999px; overflow: hidden; margin-top: 10px;
    }
    .progress-fill { height: 100%; border-radius: 999px; transition: width 0.4s ease; }
    .fill-ram { background: linear-gradient(90deg, #10b981, #00f0ff); }

    .bots-grid {
      display: grid; grid-template-columns: repeat(auto-fill, minmax(380px, 1fr)); gap: 20px; margin-bottom: 32px;
    }
    .bot-card {
      background: var(--card-bg); backdrop-filter: blur(12px); border-radius: 18px; padding: 22px;
      display: flex; flex-direction: column; justify-content: space-between; transition: all 0.25s ease;
      border: 1px solid var(--theme-border, var(--card-border));
      border-top: 4px solid var(--theme, var(--accent));
    }
    .bot-card:hover {
      box-shadow: 0 12px 36px -12px var(--theme-glow, rgba(0, 240, 255, 0.25));
    }
    .bot-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 14px; }
    .bot-name-group h3 { font-size: 1.05rem; font-weight: 700; margin-bottom: 3px; }
    .bot-port {
      font-family: 'JetBrains Mono', monospace; font-size: 0.68rem; padding: 2px 8px;
      border-radius: 6px; background: rgba(255, 255, 255, 0.05); color: var(--text-muted);
    }
    .badge-status {
      font-family: 'JetBrains Mono', monospace; font-size: 0.7rem; font-weight: 700;
      padding: 3px 10px; border-radius: 999px; border: 1px solid var(--theme-border, var(--card-border));
      background: var(--theme-bg, rgba(255,255,255,0.05)); color: var(--theme, var(--text));
    }
    .acc-list { display: flex; flex-direction: column; gap: 10px; margin: 14px 0; }
    .acc-item {
      background: rgba(0, 0, 0, 0.35); border: 1px solid rgba(255, 255, 255, 0.04);
      border-radius: 10px; padding: 12px; font-size: 0.8rem;
    }
    .acc-item-header { display: flex; justify-content: space-between; margin-bottom: 6px; font-weight: 600; }
    .stat-row { display: flex; justify-content: space-between; font-size: 0.75rem; color: var(--text-muted); margin-top: 4px; }
    .bot-footer {
      display: flex; justify-content: space-between; align-items: center; padding-top: 14px;
      border-top: 1px solid rgba(255, 255, 255, 0.06); font-size: 0.75rem;
    }
    .btn-link {
      padding: 6px 12px; border-radius: 8px; text-decoration: none; font-weight: 600;
      font-size: 0.75rem; transition: background 0.2s; border: 1px solid var(--theme-border, var(--card-border));
      color: var(--theme, var(--accent)); background: var(--theme-bg, rgba(255,255,255,0.04));
    }
    .btn-link:hover { filter: brightness(1.2); }
    .stream-ticker-wrap {
      background: rgba(15, 23, 42, 0.95); border: 1px solid var(--card-border);
      border-radius: 12px; padding: 12px 20px; display: flex; align-items: center; gap: 14px;
      overflow: hidden; margin-top: 20px;
    }
    .ticker-badge {
      background: rgba(239, 68, 68, 0.2); border: 1px solid #ef4444; color: #ef4444;
      font-size: 0.72rem; font-weight: 800; padding: 4px 10px; border-radius: 6px; white-space: nowrap;
    }
    .stream-ticker-track { display: flex; gap: 28px; white-space: nowrap; font-size: 0.8rem; color: var(--text-muted); }
  </style>
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
              <span class="bot-port">PORTA {{$bot.Meta.Port}} • {{if $bot.Meta.IsStream}}STREAM HARVESTER{{else}}GO NATIVE{{end}}</span>
            </div>
            <span class="badge-status" id="badge-{{$bot.Meta.Key}}">{{$bot.BadgeText}}</span>
          </div>

          <div class="acc-list" id="acclist-{{$bot.Meta.Key}}">
            {{if $bot.Accounts}}
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
        <span>⚡ VPS CLUSTER ONLINE • BOTS & STREAM MINERS NATIVOS EM GO • ZERO NODE.JS</span>
        <span>🎮 JOGOS & 📺 STREAMS: {{.TickerBotsList}}</span>
      </div>
    </div>
  </div>

  <script>
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
              isStream: true
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
            status: statusStr
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
            isStream: true
          };
        });
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

          if (badge && accounts.length > 0) {
            if (isStream) {
              badge.textContent = accounts[0].kills > 0 ? accounts[0].kills + ' AO VIVO' : 'MINERANDO';
            } else {
              badge.textContent = accounts.length > 1 ? accounts.length + '/' + accounts.length + ' ATIVAS' : (accounts[0].status ? accounts[0].status.toUpperCase() : 'ONLINE');
            }
          }
          if (accList && accounts.length > 0) {
            accList.innerHTML = accounts.map(function(a) {
              var s = '<div class="acc-item">';
              if (isStream) {
                s += '<div class="acc-item-header"><span>' + (a.name || 'Conta') + '</span><span style="font-size: 0.72rem; font-weight: 700;">STREAM VIEWER</span></div>';
                if (a.currentArea) {
                  s += '<div class="stat-row" style="padding: 4px 8px; border-radius: 6px; margin-bottom: 4px; background: rgba(255,255,255,0.05); font-weight: 700;">📡 ' + a.currentArea + '</div>';
                }
                s += '<div class="stat-row"><span style="color: #facc15; font-weight: 800;">⭐ ' + (a.gold || 0).toLocaleString() + ' pontos</span><span style="color: #38bdf8; font-weight: 700;">🔴 ' + (a.kills || 0) + ' ' + ((a.kills === 1) ? 'canal ao vivo' : 'canais ao vivo') + '</span></div>';
              } else {
                s += '<div class="acc-item-header"><span>' + (a.name || 'Conta') + ' (Nv. ' + (a.level || 1) + ')</span><span>' + (a.vocation || '') + '</span></div>';
                if (a.currentArea) {
                  s += '<div class="stat-row" style="padding: 4px 8px; border-radius: 6px; margin-bottom: 4px; background: rgba(255,255,255,0.05); font-weight: 700;">🗺️ ' + a.currentArea + '</div>';
                }
                if (a.maxHp) {
                  s += '<div class="stat-row"><span>❤️ HP</span><span style="color: #4ade80;">' + (a.hp || 0).toLocaleString() + ' / ' + (a.maxHp || 0).toLocaleString() + '</span></div>';
                }
                if (a.stamina) {
                  s += '<div class="stat-row"><span>⚡ Stamina</span><span style="color: #38bdf8; font-weight: 600;">' + a.stamina + '</span></div>';
                }
                s += '<div class="stat-row"><span style="color: #facc15; font-weight: 800;">💰 ' + (a.gold || 0).toLocaleString() + ' gp</span><span style="color: #38bdf8;">⚔️ ' + (a.kills || 0).toLocaleString() + ' kills</span></div>';
              }
              s += '</div>';
              return s;
            }).join('');
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
