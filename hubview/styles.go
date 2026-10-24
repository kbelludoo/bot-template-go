package hubview

// hubStyles contains all modern, hardware-accelerated CSS3 animations and styling
const hubStyles = `
    :root {
      --bg: #07090e;
      --card-bg: rgba(15, 23, 42, 0.75);
      --card-border: rgba(255, 255, 255, 0.08);
      --card-hover: rgba(255, 255, 255, 0.16);
      --accent: #00f0ff;
      --accent-glow: rgba(0, 240, 255, 0.25);
      --green: #10b981;
      --green-glow: rgba(16, 185, 129, 0.35);
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
        radial-gradient(circle at 10% 20%, rgba(0, 240, 255, 0.06) 0%, transparent 40%),
        radial-gradient(circle at 90% 10%, rgba(168, 85, 247, 0.07) 0%, transparent 45%),
        radial-gradient(circle at 50% 80%, rgba(16, 185, 129, 0.05) 0%, transparent 50%),
        linear-gradient(to bottom, #07090e, #0a0e1a);
      background-attachment: fixed;
      color: var(--text);
      font-family: 'Plus Jakarta Sans', -apple-system, sans-serif;
      min-height: 100vh;
      padding: 24px;
      line-height: 1.5;
    }
    .container { max-width: 1360px; margin: 0 auto; }

    /* Header */
    .header {
      display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 16px;
      padding: 22px 28px; background: var(--card-bg); backdrop-filter: blur(16px);
      border: 1px solid var(--card-border); border-radius: 20px; margin-bottom: 24px;
      box-shadow: 0 10px 35px -10px rgba(0, 0, 0, 0.6);
      animation: fadeIn 0.4s ease-out;
    }
    .header-left { display: flex; align-items: center; gap: 16px; }
    .logo-badge {
      width: 52px; height: 52px; border-radius: 14px;
      background: linear-gradient(135deg, rgba(0, 240, 255, 0.2), rgba(139, 92, 246, 0.3));
      border: 1px solid var(--accent); display: flex; align-items: center; justify-content: center;
      font-size: 1.7rem; box-shadow: 0 0 20px var(--accent-glow);
      animation: logoPulse 4s ease-in-out infinite alternate;
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
    
    /* Live Radar Pulse */
    .live-badge {
      display: inline-flex; align-items: center; gap: 9px; padding: 7px 16px; border-radius: 999px;
      font-size: 0.78rem; font-weight: 700; font-family: 'JetBrains Mono', monospace;
      background: rgba(16, 185, 129, 0.12); border: 1px solid var(--green); color: var(--green);
      box-shadow: 0 0 15px rgba(16, 185, 129, 0.15);
    }
    .pulse-dot {
      position: relative;
      width: 8px; height: 8px; background: var(--green); border-radius: 50%;
      box-shadow: 0 0 8px var(--green);
    }
    .pulse-dot::after {
      content: '';
      position: absolute;
      top: -3px; left: -3px; right: -3px; bottom: -3px;
      border-radius: 50%;
      border: 1.5px solid var(--green);
      animation: radarPing 2.2s cubic-bezier(0, 0, 0.2, 1) infinite;
    }

    /* VPS Hardware Metrics Grid */
    .vps-metrics-grid {
      display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 16px; margin-bottom: 28px;
    }
    .vps-card {
      background: var(--card-bg); backdrop-filter: blur(12px); border: 1px solid var(--card-border);
      border-radius: 16px; padding: 18px 22px;
      transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), border-color 0.25s ease, box-shadow 0.25s ease;
    }
    .vps-card:hover {
      transform: translateY(-3px);
      border-color: var(--card-hover);
      box-shadow: 0 10px 25px -8px rgba(0, 0, 0, 0.5);
    }
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
    .progress-fill { height: 100%; border-radius: 999px; transition: width 0.6s ease; }
    .fill-ram {
      background: linear-gradient(90deg, #10b981, #00f0ff);
      position: relative;
      overflow: hidden;
    }
    .fill-ram::after {
      content: '';
      position: absolute; top: 0; left: -100%; width: 100%; height: 100%;
      background: linear-gradient(90deg, transparent, rgba(255,255,255,0.4), transparent);
      animation: barShimmer 3s infinite;
    }

    /* Cards Grid & Micro-interactions */
    .bots-grid {
      display: grid; grid-template-columns: repeat(auto-fill, minmax(380px, 1fr)); gap: 20px; margin-bottom: 32px;
    }
    .bot-card {
      background: var(--card-bg); backdrop-filter: blur(12px); border-radius: 18px; padding: 22px;
      display: flex; flex-direction: column; justify-content: space-between;
      border: 1px solid var(--theme-border, var(--card-border));
      border-top: 4px solid var(--theme, var(--accent));
      transition: transform 0.28s cubic-bezier(0.16, 1, 0.3, 1),
                  box-shadow 0.28s cubic-bezier(0.16, 1, 0.3, 1),
                  border-color 0.28s ease;
      will-change: transform, box-shadow;
      animation: cardEntrance 0.5s cubic-bezier(0.16, 1, 0.3, 1) backwards;
    }
    .bot-card:hover {
      transform: translateY(-5px);
      box-shadow: 0 16px 40px -12px var(--theme-glow, rgba(0, 240, 255, 0.35)),
                  0 0 20px -3px var(--theme-border, rgba(0, 240, 255, 0.25));
      border-color: var(--theme, var(--accent));
    }
    .bot-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 14px; }
    .bot-name-group h3 { font-size: 1.05rem; font-weight: 700; margin-bottom: 3px; }
    .bot-port {
      font-family: 'JetBrains Mono', monospace; font-size: 0.68rem; padding: 2px 8px;
      border-radius: 6px; background: rgba(255, 255, 255, 0.05); color: var(--text-muted);
    }
    
    /* Dual Face Controls & Views */
    .face-header-controls { display: flex; align-items: center; gap: 8px; }
    .face-toggle-btn {
      background: rgba(255, 255, 255, 0.05);
      border: 1px solid var(--theme-border, var(--card-border));
      border-radius: 6px;
      color: var(--theme, var(--text-muted));
      font-family: 'JetBrains Mono', monospace;
      font-size: 0.65rem; font-weight: 700; padding: 3px 8px;
      cursor: pointer; display: inline-flex; align-items: center; gap: 4px;
      transition: all 0.2s ease; user-select: none;
    }
    .face-toggle-btn:hover {
      background: var(--theme-bg, rgba(255,255,255,0.12));
      filter: brightness(1.2); transform: scale(1.05);
    }
    .card-face {
      transition: opacity 0.3s ease, transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
      will-change: opacity, transform;
    }
    .face-performance {
      display: none; opacity: 0; transform: translateY(6px);
    }
    .bot-card.is-flipped .face-gameplay {
      display: none; opacity: 0;
    }
    .bot-card.is-flipped .face-performance {
      display: block; opacity: 1; transform: translateY(0);
    }
    .perf-tag {
      font-family: 'JetBrains Mono', monospace; font-size: 0.7rem; font-weight: 700;
      padding: 2px 6px; border-radius: 4px;
      background: rgba(16, 185, 129, 0.1); border: 1px solid rgba(16, 185, 129, 0.3); color: var(--green);
    }

    .badge-status {
      font-family: 'JetBrains Mono', monospace; font-size: 0.7rem; font-weight: 700;
      padding: 3px 10px; border-radius: 999px; border: 1px solid var(--theme-border, var(--card-border));
      background: var(--theme-bg, rgba(255,255,255,0.05)); color: var(--theme, var(--text));
      position: relative; overflow: hidden;
      display: inline-flex; align-items: center; gap: 4px;
    }
    .badge-status::before {
      content: ''; position: absolute; top: 0; left: -100%; width: 100%; height: 100%;
      background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
      animation: badgeShine 4.5s infinite;
    }
    .acc-list { display: flex; flex-direction: column; gap: 10px; margin: 14px 0; }
    .acc-item {
      background: rgba(0, 0, 0, 0.35); border: 1px solid rgba(255, 255, 255, 0.04);
      border-radius: 10px; padding: 12px; font-size: 0.8rem;
      transition: background 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
    }
    .acc-item:hover {
      background: rgba(255, 255, 255, 0.04);
      border-color: rgba(255, 255, 255, 0.1);
      transform: translateX(3px);
    }
    .acc-item-header { display: flex; justify-content: space-between; margin-bottom: 6px; font-weight: 600; }
    .stat-row { display: flex; justify-content: space-between; font-size: 0.75rem; color: var(--text-muted); margin-top: 4px; }
    .bot-footer {
      display: flex; justify-content: space-between; align-items: center; padding-top: 14px;
      border-top: 1px solid rgba(255, 255, 255, 0.06); font-size: 0.75rem;
    }
    .btn-link {
      padding: 6px 12px; border-radius: 8px; text-decoration: none; font-weight: 600;
      font-size: 0.75rem; border: 1px solid var(--theme-border, var(--card-border));
      color: var(--theme, var(--accent)); background: var(--theme-bg, rgba(255,255,255,0.04));
      transition: transform 0.2s ease, filter 0.2s ease;
      will-change: transform;
    }
    .btn-link:hover {
      transform: scale(1.04); filter: brightness(1.25);
    }

    /* Live Audio Wave Equalizer */
    .live-wave {
      display: inline-flex; align-items: flex-end; gap: 2px; height: 11px; margin-right: 2px;
    }
    .live-wave span {
      display: block; width: 2.5px; background: currentColor; border-radius: 1px;
      animation: liveWave 1.1s ease-in-out infinite alternate;
    }
    .live-wave span:nth-child(1) { height: 35%; animation-delay: 0.1s; }
    .live-wave span:nth-child(2) { height: 95%; animation-delay: 0.35s; }
    .live-wave span:nth-child(3) { height: 60%; animation-delay: 0.2s; }

    /* Stream Ticker Marquee */
    .stream-ticker-wrap {
      background: rgba(15, 23, 42, 0.95); border: 1px solid var(--card-border);
      border-radius: 12px; padding: 12px 20px; display: flex; align-items: center; gap: 14px;
      overflow: hidden; margin-top: 20px; position: relative;
    }
    .stream-ticker-wrap::before, .stream-ticker-wrap::after {
      content: ''; position: absolute; top: 0; bottom: 0; width: 45px; z-index: 2; pointer-events: none;
    }
    .stream-ticker-wrap::before { left: 0; background: linear-gradient(90deg, rgba(15,23,42,0.95), transparent); }
    .stream-ticker-wrap::after { right: 0; background: linear-gradient(270deg, rgba(15,23,42,0.95), transparent); }
    .ticker-badge {
      background: rgba(239, 68, 68, 0.2); border: 1px solid #ef4444; color: #ef4444;
      font-size: 0.72rem; font-weight: 800; padding: 4px 10px; border-radius: 6px; white-space: nowrap;
      z-index: 3;
    }
    .stream-ticker-track {
      display: flex; gap: 36px; white-space: nowrap; font-size: 0.8rem; color: var(--text-muted);
      animation: tickerScroll 40s linear infinite;
    }
    .stream-ticker-track:hover { animation-play-state: paused; }

    /* Keyframes */
    @keyframes radarPing {
      0% { transform: scale(0.8); opacity: 0.9; }
      75%, 100% { transform: scale(2.6); opacity: 0; }
    }
    @keyframes logoPulse {
      0% { box-shadow: 0 0 15px var(--accent-glow); transform: scale(1); }
      100% { box-shadow: 0 0 26px rgba(139, 92, 246, 0.4); transform: scale(1.02); }
    }
    @keyframes cardEntrance {
      from { opacity: 0; transform: translateY(14px); }
      to { opacity: 1; transform: translateY(0); }
    }
    @keyframes badgeShine {
      0%, 75% { left: -100%; }
      100% { left: 100%; }
    }
    @keyframes barShimmer {
      0% { left: -100%; }
      50%, 100% { left: 100%; }
    }
    @keyframes liveWave {
      0% { height: 20%; }
      100% { height: 100%; }
    }
    @keyframes tickerScroll {
      0% { transform: translate3d(0, 0, 0); }
      100% { transform: translate3d(-50%, 0, 0); }
    }
    @keyframes fadeIn {
      from { opacity: 0; transform: translateY(-8px); }
      to { opacity: 1; transform: translateY(0); }
    }
`
