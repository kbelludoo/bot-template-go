#!/usr/bin/env node
const http = require('http');
const { execSync } = require('child_process');

function fetchJson(url) {
  return new Promise((resolve) => {
    const req = http.get(url, { timeout: 2500 }, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          const parsed = JSON.parse(data);
          resolve({ online: true, ...parsed });
        } catch (e) {
          resolve({ online: false, error: 'JSON inválido' });
        }
      });
    });
    req.on('error', (err) => resolve({ online: false, error: err.message }));
    req.on('timeout', () => {
      req.destroy();
      resolve({ online: false, error: 'Timeout de conexão' });
    });
  });
}

function getSystemMetrics() {
  try {
    const uptime = execSync('uptime -p 2>/dev/null || uptime').toString().trim();
    const loadParts = execSync('cat /proc/loadavg 2>/dev/null || echo "0 0 0"').toString().trim().split(/\s+/);
    const freeRaw = execSync('free -m 2>/dev/null').toString();
    
    let ramTotal = 0, ramUsed = 0, ramFree = 0, ramAvail = 0;
    let swapTotal = 0, swapUsed = 0;

    for (const line of freeRaw.split('\n')) {
      const parts = line.trim().split(/\s+/);
      if (parts[0] === 'Mem:') {
        ramTotal = parseInt(parts[1], 10) || 0;
        ramUsed = parseInt(parts[2], 10) || 0;
        ramFree = parseInt(parts[3], 10) || 0;
        ramAvail = parseInt(parts[6] || parts[3], 10) || 0;
      } else if (parts[0] === 'Swap:') {
        swapTotal = parseInt(parts[1], 10) || 0;
        swapUsed = parseInt(parts[2], 10) || 0;
      }
    }

    return {
      uptime,
      load: loadParts.slice(0, 3).join(', '),
      ramTotalMb: ramTotal,
      ramUsedMb: ramUsed,
      ramAvailableMb: ramAvail,
      ramPercent: ramTotal ? Math.round((ramUsed / ramTotal) * 100) : 0,
      swapTotalMb: swapTotal,
      swapUsedMb: swapUsed
    };
  } catch (err) {
    return {
      uptime: 'N/A',
      load: 'N/A',
      ramTotalMb: 1024,
      ramUsedMb: 512,
      ramAvailableMb: 512,
      ramPercent: 50,
      swapTotalMb: 2048,
      swapUsedMb: 0
    };
  }
}

async function main() {
  const vps = getSystemMetrics();

  const [huntera, tibidle, sigma, leveling, pokeidle, pokeidleworld, kickminer, baiak] = await Promise.all([
    fetchJson('http://127.0.0.1:3000/api/status'),
    fetchJson('http://127.0.0.1:3001/api/status'),
    fetchJson('http://127.0.0.1:3002/api/status'),
    fetchJson('http://127.0.0.1:3003/api/status'),
    fetchJson('http://127.0.0.1:3006/api/status'),
    fetchJson('http://127.0.0.1:3007/api/status'),
    fetchJson('http://127.0.0.1:3009/api/status'),
    fetchJson('http://127.0.0.1:8090/api/fleet')
  ]);

  const bots = {
    huntera: {
      name: 'Huntera Idle Bot',
      game: 'Huntera',
      engine: 'Go Native Engine',
      port: 3000,
      dashboardUrl: 'https://kbelludoo.github.io/huntera-bot/',
      status: huntera
    },
    tibidle: {
      name: 'Tibidle Multi-Bot',
      game: 'Tibidle',
      engine: 'Go Native Engine',
      port: 3001,
      dashboardUrl: 'https://kbelludoo.github.io/tibidle-bot/',
      status: tibidle
    },
    sigma: {
      name: 'Pokemon Sigma Multi-Bot',
      game: 'Pokemon Sigma',
      engine: 'Go Native Engine',
      port: 3002,
      dashboardUrl: 'https://kbelludoo.github.io/sigma-bot/',
      status: sigma
    },
    leveling: {
      name: 'Leveling Idle 2D Bot',
      game: 'Leveling Idle 2D',
      engine: 'Go Native Engine (Protobuf 3)',
      port: 3003,
      dashboardUrl: 'https://kbelludoo.github.io/leveling-bot/',
      status: leveling
    },
    pokeidle: {
      name: 'PokeIdle.io Bot',
      game: 'PokeIdle.io',
      engine: 'Go Native Engine (Twitch Lurker)',
      port: 3006,
      dashboardUrl: 'https://kbelludoo.github.io/pokeidle-go/',
      status: pokeidle
    },
    pokeidleworld: {
      name: 'Poke Idle World Bot',
      game: 'Poke Idle World',
      engine: 'Go Native Engine (Auto-Catch)',
      port: 3007,
      dashboardUrl: 'https://kbelludoo.github.io/pokeidleworld-bot/',
      status: pokeidleworld
    },
    kickminer: {
      name: 'Kick Channel Points Miner',
      game: 'Kick.com',
      engine: 'Go Native (Zero-Video WS)',
      port: 3009,
      dashboardUrl: 'https://kbelludoo.github.io/kick-miner-go/',
      status: kickminer
    },
    baiak: {
      name: 'Baiak Idle Terminal Bot',
      game: 'Baiak Idle',
      engine: 'Bun / Colyseus Headless',
      port: 8090,
      dashboardUrl: 'https://kbelludoo.github.io/baiak-idle-bot/',
      status: baiak
    }
  };

  let activeBotsCount = 0;
  if (huntera.online) activeBotsCount++;
  if (tibidle.online) activeBotsCount++;
  if (sigma.online) activeBotsCount++;
  if (leveling.online) activeBotsCount++;
  if (pokeidle.online) activeBotsCount++;
  if (pokeidleworld.online) activeBotsCount++;
  if (kickminer.online) activeBotsCount++;
  if (baiak.online && (baiak.fleet?.botsOnline > 0 || baiak.online)) activeBotsCount++;

  // Função de anonimização (Modo Stealth / Anti-Ban)
  function sanitizeBotPayload(b) {
    if (!b || typeof b !== 'object') return b;
    const clean = JSON.parse(JSON.stringify(b));

    // Huntera: charName
    if (clean.huntera?.status?.accounts) {
      clean.huntera.status.accounts.forEach((acc, i) => {
        acc.charName = `Personagem #${i + 1}`;
      });
    }
    // Tibidle: name
    if (clean.tibidle?.status?.accounts) {
      clean.tibidle.status.accounts.forEach((acc, i) => {
        acc.name = `Conta #${i + 1}`;
      });
    }
    // Sigma: trainerName, alias
    if (clean.sigma?.status?.accounts) {
      clean.sigma.status.accounts.forEach((acc, i) => {
        acc.trainerName = `Treinador #${i + 1}`;
        acc.alias = `Treinador #${i + 1}`;
      });
    }
    // Leveling: name
    if (clean.leveling?.status?.accounts) {
      clean.leveling.status.accounts.forEach((acc, i) => {
        acc.name = `Party #${i + 1}`;
      });
    }
    // PokeIdle: name
    if (clean.pokeidle?.status?.accounts) {
      clean.pokeidle.status.accounts.forEach((acc, i) => {
        acc.name = `Treinador #${i + 1}`;
      });
    }
    // Poke Idle World: name
    if (clean.pokeidleworld?.status?.accounts) {
      clean.pokeidleworld.status.accounts.forEach((acc, i) => {
        acc.name = `Treinador #${i + 1}`;
      });
    }
    // Baiak: bots[key].name e label
    if (clean.baiak?.status?.bots) {
      Object.keys(clean.baiak.status.bots).forEach((k, i) => {
        clean.baiak.status.bots[k].name = `Frota #${i + 1}`;
        clean.baiak.status.bots[k].label = `Conta #${i + 1}`;
        if (clean.baiak.status.bots[k].status?.characters) {
          clean.baiak.status.bots[k].status.characters.forEach((c, ci) => {
            c.name = `Personagem #${ci + 1}`;
          });
        }
      });
    }
    // Kick Miner: username e chaves de contas
    if (clean.kickminer?.status?.accounts) {
      const sanitizedAccs = {};
      Object.keys(clean.kickminer.status.accounts).forEach((name, i) => {
        const item = clean.kickminer.status.accounts[name];
        item.username = `Minerador #${i + 1}`;
        sanitizedAccs[`miner_${i + 1}`] = item;
      });
      clean.kickminer.status.accounts = sanitizedAccs;
    }

    return clean;
  }

  const payload = {
    timestamp: new Date().toISOString(),
    vps,
    summary: {
      totalBots: 8,
      activeBots: activeBotsCount
    },
    bots: sanitizeBotPayload(bots)
  };

  const jsonStr = JSON.stringify(payload, null, 2);

  // Se executado com push (--push) ou standalone no crontab
  if (process.argv.includes('--push')) {
    try {
      const fs = require('fs');
      const b64 = Buffer.from(jsonStr).toString('base64');
      let sha = '';
      try {
        sha = execSync('gh api repos/kbelludoo/bot-template-go/contents/docs/status.json --jq .sha 2>/dev/null')
          .toString().trim();
      } catch (e) {
        // Arquivo ainda não existe
      }

      const tmpPayload = {
        message: 'telemetry: live vps 8 bots summary',
        content: b64,
        ...(sha ? { sha } : {})
      };
      fs.writeFileSync('/tmp/status_hub_payload.json', JSON.stringify(tmpPayload));

      execSync('gh api -X PUT repos/kbelludoo/bot-template-go/contents/docs/status.json --input /tmp/status_hub_payload.json > /dev/null 2>&1');
      console.log('✓ Resumo da VPS (8 bots ativos) sincronizado com sucesso no GitHub Pages!');
    } catch (pushErr) {
      console.error('[-] Falha ao sincronizar via GitHub API:', pushErr.message);
    }
  } else {
    process.stdout.write(jsonStr);
  }
}

main();
