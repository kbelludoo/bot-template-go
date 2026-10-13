#!/usr/bin/env node
const http = require('http');
const { execSync } = require('child_process');
const fs = require('fs');

// MODO STEALTH: scrub profundo aplicado a TODO payload antes do push.
// Remove nicks, e-mails e IPs de qualquer campo (inclusive logs e chaves),
// preservando apenas URLs http(s) (dashboards/túneis).
const NICK_RE = /kbelludoo[0-9]*|lohmoreira|\blrn\b/gi;
const EMAIL_RE = /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/gi;
const IP_RE = /\b\d{1,3}(?:\.\d{1,3}){3}\b/g;

function stealthScrub(value) {
  if (typeof value === 'string') {
    if (/^https?:\/\//i.test(value)) return value;
    return value
      .replace(/\([^)]*@[^)]*\)/g, '(Sessão Autenticada)')
      .replace(EMAIL_RE, 'protegido')
      .replace(IP_RE, '127.0.0.1')
      .replace(NICK_RE, 'treinador');
  }
  if (Array.isArray(value)) return value.map(stealthScrub);
  if (value && typeof value === 'object') {
    const out = {};
    let n = 0;
    for (const rawKey of Object.keys(value)) {
      let key = rawKey;
      if (!/^https?:\/\//i.test(rawKey)) key = rawKey.replace(NICK_RE, 'treinador');
      while (Object.prototype.hasOwnProperty.call(out, key)) key = `${key}_${++n}`;
      out[key] = stealthScrub(value[rawKey]);
    }
    return out;
  }
  return value;
}

// Máscara por bot: substitui nicks reais por rótulos anônimos por conta.
const BOTS = [
  { repo: 'sigma-bot', port: 3002, label: 'Treinador', fields: ['trainerName', 'alias'] },
  { repo: 'tibidle-bot', port: 3001, label: 'Conta', fields: ['name'] },
  { repo: 'leveling-bot', port: 3003, label: 'Party', fields: ['name'] },
  { repo: 'pokeidle-go', port: 3006, label: 'Treinador', fields: ['name'] },
  { repo: 'pokeidleworld-bot', port: 3007, label: 'Treinador', fields: ['name'] }
];

function fetchJson(url) {
  return new Promise((resolve) => {
    http.get(url, { timeout: 4000 }, (res) => {
      let data = '';
      res.on('data', c => data += c);
      res.on('end', () => {
        try { resolve(JSON.parse(data)); } catch { resolve(null); }
      });
    }).on('error', () => resolve(null));
  });
}

function anonymizeAccounts(payload, bot) {
  if (!payload || !Array.isArray(payload.accounts)) return payload;
  payload.accounts.forEach((acc, i) => {
    const label = `${bot.label} #${i + 1}`;
    bot.fields.forEach(f => { if (typeof acc[f] === 'string') acc[f] = label; });
    if (typeof acc.id === 'string') acc.id = `acc-${i + 1}`;
    if (acc.extra && typeof acc.extra.trainerName === 'string') acc.extra.trainerName = label;
    ['logs', 'recentLogs'].forEach(k => {
      if (Array.isArray(acc[k])) {
        acc[k] = acc[k].map(l => {
          if (typeof l !== 'string') return l;
          return l
            .replace(/\([^)]*@[^)]*\)/g, '(Sessão Autenticada)')
            .replace(EMAIL_RE, 'protegido')
            .replace(NICK_RE, label);
        });
      }
    });
    if (acc.bindIp) acc.bindIp = '127.0.0.1';
    if (acc.outboundIp) acc.outboundIp = '127.0.0.1';
  });
  return payload;
}

function pushToGh(repo, jsonPayload) {
  try {
    const safe = stealthScrub(jsonPayload);
    const b64 = Buffer.from(JSON.stringify(safe, null, 2)).toString('base64');
    let sha = '';
    try {
      sha = execSync(`gh api repos/kbelludoo/${repo}/contents/docs/status.json --jq .sha 2>/dev/null`, { timeout: 10000 }).toString().trim();
    } catch {}

    const tmpPath = `/tmp/status_${repo}.json`;
    const body = {
      message: 'telemetry: live bot status update (stealth)',
      content: b64,
      ...(sha ? { sha } : {})
    };
    fs.writeFileSync(tmpPath, JSON.stringify(body));
    execSync(`gh api -X PUT repos/kbelludoo/${repo}/contents/docs/status.json --input ${tmpPath} >/dev/null 2>&1`, { timeout: 15000 });
    console.log(`✓ [${repo}] Telemetria anonimizada sincronizada com sucesso!`);
  } catch (err) {
    console.error(`[-] [${repo}] Erro ao sincronizar:`, err.message);
  }
}

// Garante que huntera-bot docs/index.html não contenha IP do host
function sanitizeHunteraIndex() {
  try {
    const repo = 'huntera-bot';
    const sha = execSync(`gh api repos/kbelludoo/${repo}/contents/docs/index.html --jq .sha 2>/dev/null`, { timeout: 10000 }).toString().trim();
    if (!sha) return;
    const rawB64 = execSync(`gh api repos/kbelludoo/${repo}/contents/docs/index.html --jq .content 2>/dev/null`, { timeout: 10000 }).toString().trim();
    const raw = Buffer.from(rawB64, 'base64').toString('utf8');

    if (raw.includes('129.153.157.17')) {
      const sanitized = raw.split('129.153.157.17').join('127.0.0.1');
      const b64 = Buffer.from(sanitized).toString('base64');
      fs.writeFileSync('/tmp/huntera_index.json', JSON.stringify({
        message: 'security: sanitize raw VPS IP from index template',
        content: b64,
        sha: sha
      }));
      execSync(`gh api -X PUT repos/kbelludoo/${repo}/contents/docs/index.html --input ${tmpPath} >/dev/null 2>&1`, { timeout: 15000 });
      console.log(`✓ [${repo}] docs/index.html sanitizado (IP removido)!`);
    }
  } catch (err) {
    // ignorar
  }
}

async function syncAll() {
  sanitizeHunteraIndex();
  for (const bot of BOTS) {
    const status = await fetchJson(`http://127.0.0.1:${bot.port}/api/status`);
    if (!status) continue;
    pushToGh(bot.repo, anonymizeAccounts(status, bot));
  }
}

syncAll();
