#!/usr/bin/env node
/**
 * watchdog-agent.js
 * 
 * Monitor Central de Saúde dos 8 Bots na VPS.
 * Ciclo de Operação:
 *  1. Verifica a saúde de cada bot a cada minuto.
 *  2. Se detectar bot offline ou sem responder na porta:
 *     a. Aciona auto-cura local (systemctl restart ou reinício de processo).
 *     b. Aciona o Agent em Go (porta 8888) para diagnóstico inteligente com LLM.
 *     c. Aguarda até 10 segundos e re-testa.
 *  3. Se a auto-cura resolver:
 *     - Envia notificação informativa no Telegram de recuperação bem-sucedida.
 *  4. Se a auto-cura NÃO resolver:
 *     - Envia alerta crítico com o relatório e causa provável diagnosticada pelo Agent em Go.
 */

const http = require('http');
const https = require('https');
const { execSync } = require('child_process');

const TELEGRAM_TOKEN = '8971166553:AAEyGA3iAdHZzIbUNB8ejzPRnNVDt6oY5aA';
const TELEGRAM_CHAT_ID = '1129361331';

const BOTS_CONFIG = [
  {
    id: 'huntera',
    name: '⚔️ Huntera Idle',
    port: 3000,
    checkUrl: 'http://127.0.0.1:3000/api/status',
    restartCmd: 'bash /home/ubuntu/huntera-go/start.sh',
    type: 'tmux'
  },
  {
    id: 'tibidle',
    name: '🛡️ Tibidle Multi-Bot',
    port: 3001,
    checkUrl: 'http://127.0.0.1:3001/api/status',
    restartCmd: 'sudo systemctl restart tibidle.service',
    type: 'systemd'
  },
  {
    id: 'sigma',
    name: '⚡ Pokemon Sigma',
    port: 3002,
    checkUrl: 'http://127.0.0.1:3002/api/status',
    restartCmd: 'sudo systemctl restart sigma-bot.service',
    type: 'systemd'
  },
  {
    id: 'leveling',
    name: '🏹 Leveling Idle 2D',
    port: 3003,
    checkUrl: 'http://127.0.0.1:3003/api/status',
    restartCmd: 'sudo systemctl restart leveling-bot.service',
    type: 'systemd'
  },
  {
    id: 'pokeidle',
    name: '⚡ PokeIdle.io',
    port: 3006,
    checkUrl: 'http://127.0.0.1:3006/api/status',
    restartCmd: 'sudo systemctl restart pokeidle.service',
    type: 'systemd'
  },
  {
    id: 'pokeidleworld',
    name: '🌍 Poke Idle World',
    port: 3007,
    checkUrl: 'http://127.0.0.1:3007/api/status',
    restartCmd: 'sudo systemctl restart pokeidleworld.service',
    type: 'systemd'
  },
  {
    id: 'kickminer',
    name: '🟢 Kick Miner Go',
    port: 3009,
    checkUrl: 'http://127.0.0.1:3009/api/status',
    restartCmd: 'sudo systemctl restart kick-miner.service',
    type: 'systemd'
  },
  {
    id: 'twitchminer',
    name: '🟣 Twitch Miner Go',
    port: 3010,
    checkUrl: 'http://127.0.0.1:3010/api/status',
    restartCmd: 'sudo systemctl restart twitch-miner.service',
    type: 'systemd'
  },
  {
    id: 'baiak',
    name: '🏰 Baiak Idle (Fleet)',
    port: 8090,
    checkUrl: 'http://127.0.0.1:8090/api/fleet',
    restartCmd: 'sudo systemctl restart baiak-fleet.service baiak-term.service baiak-term-acc2.service baiak-term-acc3.service',
    type: 'systemd'
  }
];

function sendTelegram(text) {
  return new Promise((resolve) => {
    try {
      const payload = JSON.stringify({
        chat_id: TELEGRAM_CHAT_ID,
        text: text,
        parse_mode: 'Markdown'
      });

      const req = https.request({
        hostname: 'api.telegram.org',
        path: `/bot${TELEGRAM_TOKEN}/sendMessage`,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json; charset=utf-8',
          'Content-Length': Buffer.byteLength(payload)
        },
        timeout: 8000
      }, (res) => {
        let respData = '';
        res.on('data', chunk => respData += chunk);
        res.on('end', () => resolve(true));
      });

      req.on('error', (e) => {
        console.error('[-] Falha ao enviar telegram:', e.message);
        resolve(false);
      });
      req.on('timeout', () => {
        req.destroy();
        resolve(false);
      });
      req.write(payload);
      req.end();
    } catch (err) {
      console.error('[-] Erro sendTelegram:', err.message);
      resolve(false);
    }
  });
}

function probePort(url, timeoutMs = 3000) {
  return new Promise((resolve) => {
    const req = http.get(url, { timeout: timeoutMs }, (res) => {
      let data = '';
      res.on('data', c => data += c);
      res.on('end', () => {
        try {
          const parsed = JSON.parse(data);
          resolve({ ok: true, data: parsed });
        } catch {
          resolve({ ok: res.statusCode === 200, data: null });
        }
      });
    });
    req.on('error', (e) => resolve({ ok: false, error: e.message }));
    req.on('timeout', () => {
      req.destroy();
      resolve({ ok: false, error: 'Timeout' });
    });
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
        timeout: 45000
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
      req.on('error', () => resolve(null));
      req.on('timeout', () => {
        req.destroy();
        resolve(null);
      });
      req.write(payload);
      req.end();
    } catch {
      resolve(null);
    }
  });
}

function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}

// Histórico de incidentes para evitar flood de mensagens repetidas
const incidentHistory = {};

async function runWatchdogCycle() {
  console.log(`[watchdog] Verificando integridade dos 8 bots... (${new Date().toLocaleTimeString('pt-BR')})`);

  for (const bot of BOTS_CONFIG) {
    const probe = await probePort(bot.checkUrl, 3000);

    let isHealthy = probe.ok;
    // Checagem específica para Baiak (se o fleet estiver vazio ou com 0 contas online)
    if (bot.id === 'baiak' && probe.ok && probe.data) {
      if (probe.data.fleet && probe.data.fleet.botsOnline === 0) {
        isHealthy = false;
      }
    }

    if (isHealthy) {
      if (incidentHistory[bot.id]) {
        console.log(`[watchdog] Bot ${bot.name} normalizou.`);
        delete incidentHistory[bot.id];
      }
      continue;
    }

    console.warn(`[watchdog] ⚠️ ALERTA: ${bot.name} (Porta ${bot.port}) está FORA DO AR! (${probe.error || 'sem resposta'})`);

    // 1. TENTATIVA DE AUTO-CURA
    console.log(`[watchdog] 🔧 Executando auto-cura para ${bot.name}: ${bot.restartCmd}`);
    try {
      execSync(bot.restartCmd, { timeout: 15000, stdio: 'ignore' });
    } catch (e) {
      console.error(`[-] Falha ao executar comando de auto-cura para ${bot.name}:`, e.message);
    }

    // Aguarda o processo subir e abrir porta
    await sleep(6000);

    // 2. RE-TESTE APÓS AUTO-CURA
    const retest = await probePort(bot.checkUrl, 4000);

    if (retest.ok) {
      console.log(`[watchdog] ✅ Auto-cura bem-sucedida! ${bot.name} voltou a responder na porta ${bot.port}.`);
      
      // Avisa no Telegram que o bot caiu mas foi curado sozinho
      await sendTelegram(
        `🛠️ *AUTO-CURA CONCLUÍDA NA VPS*\n\n` +
        `• *Bot:* ${bot.name}\n` +
        `• *Porta:* \`${bot.port}\`\n` +
        `• *Ação:* Serviço reiniciado automaticamente pelo Watchdog.\n` +
        `• *Status Atual:* ✅ ONLINE e operando normalmente!`
      );
      delete incidentHistory[bot.id];
    } else {
      console.error(`[watchdog] ❌ Auto-cura falhou para ${bot.name}. Acionando Agent Go para diagnóstico...`);

      // 3. SE FALHAR: ACIONA O AGENT EM GO PARA INVESTIGAR
      const agentPrompt = `O bot ${bot.name} na porta ${bot.port} caiu e a reinicializacao rapida falhou. Verifique os ultimos logs do systemd ou processos, explique por que ele nao esta subindo e o que precisa ser feito.`;
      const agentReport = await askGoAgent(agentPrompt);

      const reportClean = agentReport 
        ? agentReport.slice(0, 800) 
        : `Erro ao obter resposta do Agent Go. Verifique manualmente via SSH: \`sudo systemctl status ${bot.id}\``;

      // 4. NOTIFICAÇÃO CRÍTICA NO TELEGRAM (apenas 1 vez a cada 10 min por bot para não flodar)
      const now = Date.now();
      if (!incidentHistory[bot.id] || (now - incidentHistory[bot.id] > 600000)) {
        incidentHistory[bot.id] = now;
        await sendTelegram(
          `🚨 *ALERTA CRÍTICO: BOT FORA DO AR!*\n\n` +
          `• *Bot:* ${bot.name}\n` +
          `• *Porta:* \`${bot.port}\`\n` +
          `• *Tentativa de Auto-Cura:* ❌ Falhou ao reiniciar\n\n` +
          `🤖 *Diagnóstico do Agent Go:*\n${reportClean}`
        );
      }
    }
  }
}

// Se chamado diretamente
if (require.main === module) {
  runWatchdogCycle();
}

module.exports = { runWatchdogCycle };
