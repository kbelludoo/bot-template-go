#!/usr/bin/env node
/**
 * telegram-chat-agent.js
 * 
 * Bot interativo no Telegram para conversa bidirecional com a IA (Agent Go)
 * e controle autônomo dos 8 bots na VPS.
 * 
 * Funcionalidades:
 *  - Conversa direta com o Agent em Go (http://127.0.0.1:8888/agent) com IA (Kimi-k3/Llama-3.3)
 *  - Execução de comandos do sistema e ferramentas
 *  - Comandos rápidos nativos: /status, /restart <bot>, /help, /start
 *  - Trava de segurança: apenas o Chat ID autorizado pode interagir
 *  - Zero dependências externas (roda com Node.js puro)
 */

const http = require('http');
const https = require('https');
const { execSync } = require('child_process');

const TELEGRAM_TOKEN = '8971166553:AAEyGA3iAdHZzIbUNB8ejzPRnNVDt6oY5aA';
const ALLOWED_CHAT_ID = '1129361331';
const AGENT_URL = 'http://127.0.0.1:8888/agent';

const BOTS_MAP = {
  huntera: { name: '⚔️ Huntera Idle', port: 3000, restart: 'bash /home/ubuntu/huntera-go/start.sh' },
  tibidle: { name: '🛡️ Tibidle Multi-Bot', port: 3001, restart: 'sudo systemctl restart tibidle.service' },
  sigma: { name: '⚡ Pokemon Sigma', port: 3002, restart: 'sudo systemctl restart sigma-bot.service' },
  leveling: { name: '🏹 Leveling Idle 2D', port: 3003, restart: 'sudo systemctl restart leveling-bot.service' },
  pokeidle: { name: '⚡ PokeIdle.io', port: 3006, restart: 'sudo systemctl restart pokeidle.service' },
  pokeidleworld: { name: '🌍 Poke Idle World', port: 3007, restart: 'sudo systemctl restart pokeidleworld.service' },
  kickminer: { name: '🟢 Kick Miner Go', port: 3009, restart: 'sudo systemctl restart kick-miner.service' },
  baiak: { name: '🏰 Baiak Idle Fleet', port: 8090, restart: 'sudo systemctl restart baiak-fleet.service baiak-term.service baiak-term-acc2.service baiak-term-acc3.service' }
};

let lastUpdateId = 0;

function sendTelegramMessage(chatId, text, parseMode = 'Markdown') {
  return new Promise((resolve) => {
    // Quebra mensagens longas em pedaços de no máximo 4000 caracteres
    const chunks = [];
    let remaining = text;
    while (remaining.length > 4000) {
      let cutIndex = remaining.lastIndexOf('\n', 4000);
      if (cutIndex === -1) cutIndex = 4000;
      chunks.push(remaining.slice(0, cutIndex));
      remaining = remaining.slice(cutIndex);
    }
    chunks.push(remaining);

    (async () => {
      for (const chunk of chunks) {
        await new Promise((r) => {
          const payload = JSON.stringify({
            chat_id: chatId,
            text: chunk,
            parse_mode: parseMode
          });

          const req = https.request({
            hostname: 'api.telegram.org',
            path: `/bot${TELEGRAM_TOKEN}/sendMessage`,
            method: 'POST',
            headers: {
              'Content-Type': 'application/json; charset=utf-8',
              'Content-Length': Buffer.byteLength(payload)
            },
            timeout: 10000
          }, (res) => {
            let data = '';
            res.on('data', d => data += d);
            res.on('end', () => {
              if (res.statusCode !== 200 && parseMode === 'Markdown') {
                // Se falhar o parse Markdown do Telegram, tenta reenviar como texto simples
                sendTelegramMessage(chatId, chunk, null).then(r);
              } else {
                r();
              }
            });
          });

          req.on('error', () => r());
          req.on('timeout', () => { req.destroy(); r(); });
          req.write(payload);
          req.end();
        });
      }
      resolve();
    })();
  });
}

function sendChatAction(chatId, action = 'typing') {
  return new Promise((resolve) => {
    const payload = JSON.stringify({ chat_id: chatId, action: action });
    const req = https.request({
      hostname: 'api.telegram.org',
      path: `/bot${TELEGRAM_TOKEN}/sendChatAction`,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(payload)
      },
      timeout: 5000
    }, () => resolve());
    req.on('error', () => resolve());
    req.on('timeout', () => { req.destroy(); resolve(); });
    req.write(payload);
    req.end();
  });
}

function getUpdates(offset, timeoutSec = 25) {
  return new Promise((resolve) => {
    const path = `/bot${TELEGRAM_TOKEN}/getUpdates?offset=${offset}&timeout=${timeoutSec}`;
    const req = https.get({
      hostname: 'api.telegram.org',
      path: path,
      timeout: (timeoutSec + 10) * 1000
    }, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          const parsed = JSON.parse(data);
          resolve(parsed.ok ? parsed.result : []);
        } catch {
          resolve([]);
        }
      });
    });
    req.on('error', () => resolve([]));
    req.on('timeout', () => { req.destroy(); resolve([]); });
  });
}

function askGoAgent(prompt) {
  return new Promise((resolve) => {
    try {
      const payload = JSON.stringify({ message: prompt });
      const req = http.request({
        hostname: '127.0.0.1',
        port: 8888,
        path: '/agent',
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(payload)
        },
        timeout: 90000
      }, (res) => {
        let data = '';
        res.on('data', c => data += c);
        res.on('end', () => {
          try {
            const j = JSON.parse(data);
            resolve(j.response || j.error || data);
          } catch {
            resolve(data);
          }
        });
      });
      req.on('error', (err) => resolve(`[-] Erro ao contatar Agent Go (:8888): ${err.message}`));
      req.on('timeout', () => {
        req.destroy();
        resolve(`[-] Timeout ao aguardar resposta do Agent Go.`);
      });
      req.write(payload);
      req.end();
    } catch (e) {
      resolve(`[-] Erro interno ao chamar Agent Go: ${e.message}`);
    }
  });
}

function probePort(port) {
  return new Promise((resolve) => {
    const req = http.get(`http://127.0.0.1:${port}/api/status`, { timeout: 2000 }, (res) => {
      resolve(res.statusCode === 200);
    });
    req.on('error', () => resolve(false));
    req.on('timeout', () => { req.destroy(); resolve(false); });
  });
}

async function handleFastStatus() {
  let report = `📊 *STATUS DOS 8 BOTS NA VPS*\n\n`;
  for (const [key, bot] of Object.entries(BOTS_MAP)) {
    let online = false;
    if (key === 'baiak') {
      online = await new Promise(r => {
        const req = http.get('http://127.0.0.1:8090/api/fleet', { timeout: 2000 }, (res) => r(res.statusCode === 200));
        req.on('error', () => r(false));
        req.on('timeout', () => { req.destroy(); r(false); });
      });
    } else {
      online = await probePort(bot.port);
    }
    report += `${online ? '✅' : '❌'} *${bot.name}* (:${bot.port}) → ${online ? 'ONLINE' : 'OFFLINE'}\n`;
  }

  try {
    const freeRaw = execSync('free -m 2>/dev/null').toString();
    for (const line of freeRaw.split('\n')) {
      const parts = line.trim().split(/\s+/);
      if (parts[0] === 'Mem:') {
        report += `\n💾 *RAM:* ${parts[2]}MB usado / ${parts[1]}MB total`;
      }
    }
  } catch {}

  return report;
}

async function handleMessage(msg) {
  const chatId = String(msg.chat.id);
  const text = (msg.text || '').trim();

  // Trava de segurança: apenas usuário autorizado
  if (chatId !== ALLOWED_CHAT_ID) {
    console.warn(`[seguranca] Mensagem rejeitada de chatId nao autorizado: ${chatId}`);
    await sendTelegramMessage(chatId, `⛔ *Acesso Negado.*\nSeu Chat ID (\`${chatId}\`) não possui autorização.`, 'Markdown');
    return;
  }

  if (!text) return;

  console.log(`[telegram] Mensagem recebida: "${text}"`);

  // Comandos rápidos
  if (text === '/start') {
    const welcome = 
      `👋 *Olá, Comandante!*\n\n` +
      `Estou conectado ao seu **Agent IA em Go** e monitorando a VPS.\n\n` +
      `📌 *O que você pode fazer por aqui:*\n` +
      `• *Conversar livremente:* Digite qualquer pergunta ou instrução que a IA responderá!\n` +
      `• \`/status\` - Ver o status em tempo real dos 8 bots e da VPS\n` +
      `• \`/restart <bot>\` - Reiniciar um bot específico (ex: \`/restart sigma\`)\n` +
      `• \`/help\` - Ver lista de comandos e bots disponíveis\n\n` +
      `_Tudo pronto para operar 24/7!_ 🚀`;
    await sendTelegramMessage(chatId, welcome);
    return;
  }

  if (text === '/help') {
    const help = 
      `📖 *COMANDOS DISPONÍVEIS:*\n\n` +
      `• \`/status\` - Exibe a saúde dos 8 bots e RAM da VPS\n` +
      `• \`/restart <bot>\` - Reinicia um bot específico\n` +
      `  _Bots:_ \`huntera\`, \`tibidle\`, \`sigma\`, \`leveling\`, \`pokeidle\`, \`pokeidleworld\`, \`kickminer\`, \`baiak\`\n` +
      `• Ou envie *qualquer texto* para conversar com a IA (NVIDIA NIM / Kimi-k3 / Llama-3.3) com acesso a ferramentas do sistema!`;
    await sendTelegramMessage(chatId, help);
    return;
  }

  if (text === '/status') {
    await sendChatAction(chatId, 'typing');
    const statusMsg = await handleFastStatus();
    await sendTelegramMessage(chatId, statusMsg);
    return;
  }

  if (text.startsWith('/restart')) {
    const target = text.replace('/restart', '').trim().toLowerCase();
    const bot = BOTS_MAP[target];
    if (!bot) {
      await sendTelegramMessage(chatId, `⚠️ Bot \`${target}\` não reconhecido.\nUse um destes: \`${Object.keys(BOTS_MAP).join('`, `')}\``);
      return;
    }

    await sendChatAction(chatId, 'typing');
    await sendTelegramMessage(chatId, `🔄 Reiniciando *${bot.name}*...`);
    try {
      execSync(bot.restart, { timeout: 15000, stdio: 'ignore' });
      await sendTelegramMessage(chatId, `✅ Comando de reinício executado para *${bot.name}*. Verificando com /status em instantes.`);
    } catch (e) {
      await sendTelegramMessage(chatId, `❌ Falha ao reiniciar *${bot.name}*: ${e.message}`);
    }
    return;
  }

  // Se não for um comando pré-definido: repassa para o Agent em Go com IA!
  await sendChatAction(chatId, 'typing');
  
  // Mantém o status "digitando..." enquanto a IA gera a resposta
  const typingInterval = setInterval(() => {
    sendChatAction(chatId, 'typing');
  }, 4000);

  try {
    const aiResponse = await askGoAgent(text);
    clearInterval(typingInterval);
    await sendTelegramMessage(chatId, aiResponse || '(Sem resposta do agente)');
  } catch (err) {
    clearInterval(typingInterval);
    await sendTelegramMessage(chatId, `❌ Erro no Agent Go: ${err.message}`, null);
  }
}

async function startPolling() {
  console.log(`=======================================================`);
  console.log(`🤖 TELEGRAM CHAT AGENT INICIADO`);
  console.log(`Chat ID Autorizado: ${ALLOWED_CHAT_ID}`);
  console.log(`Conexão com Agent Go: ${AGENT_URL}`);
  console.log(`=======================================================`);

  while (true) {
    try {
      const updates = await getUpdates(lastUpdateId, 25);
      for (const update of updates) {
        lastUpdateId = update.update_id + 1;
        if (update.message) {
          await handleMessage(update.message);
        }
      }
    } catch (e) {
      console.error(`[-] Erro no loop de polling:`, e.message);
      await new Promise(r => setTimeout(r, 5000));
    }
  }
}

startPolling();
