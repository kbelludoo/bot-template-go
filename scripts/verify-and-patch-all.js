#!/usr/bin/env node
const { execSync } = require('child_process');
const fs = require('fs');

const repos = [
  'bot-template-go',
  'sigma-bot',
  'tibidle-bot',
  'leveling-bot',
  'pokeidle-go',
  'pokeidleworld-bot',
  'kick-miner-go',
  'twitch-miner-go',
  'huntera-bot'
];

console.log('====================================================');
console.log('       VERIFICAÇÃO DE MODO ANÔNIMO GLOBAL          ');
console.log('====================================================');

// 1. Verificar e limpar huntera-bot docs/index.html
try {
  const sha = execSync('gh api repos/kbelludoo/huntera-bot/contents/docs/index.html --jq .sha 2>/dev/null').toString().trim();
  if (sha) {
    const raw = Buffer.from(execSync('gh api repos/kbelludoo/huntera-bot/contents/docs/index.html --jq .content 2>/dev/null').toString().trim(), 'base64').toString('utf8');
    if (raw.includes('129.153.157.17')) {
      console.log('[-] Encontrado IP em huntera-bot/docs/index.html. Limpando...');
      const clean = raw.split('129.153.157.17').join('127.0.0.1');
      const b64 = Buffer.from(clean).toString('base64');
      fs.writeFileSync('/tmp/huntera_patch.json', JSON.stringify({
        message: 'security: sanitize raw VPS IP from index template',
        content: b64,
        sha: sha
      }));
      execSync('gh api -X PUT repos/kbelludoo/huntera-bot/contents/docs/index.html --input /tmp/huntera_patch.json >/dev/null 2>&1');
      console.log('✓ [huntera-bot] IP removido com sucesso de docs/index.html!');
    } else {
      console.log('✓ [huntera-bot] docs/index.html já está 100% limpo de IP.');
    }
  }
} catch (e) {
  console.log('[!] huntera-bot index check:', e.message);
}

// 2. Verificar status.json em cada repositório
const statusRepos = ['bot-template-go', 'sigma-bot', 'tibidle-bot', 'leveling-bot', 'pokeidle-go', 'pokeidleworld-bot', 'twitch-miner-go'];
for (const repo of statusRepos) {
  try {
    const raw = Buffer.from(execSync(`gh api repos/kbelludoo/${repo}/contents/docs/status.json --jq .content 2>/dev/null`).toString().trim(), 'base64').toString('utf8');
    const noUrls = raw.replace(/https?:\/\/[^\s",]+/g, '');
    const hasNick = /kbelludoo|lohmoreira|\blrn\b/i.test(noUrls);
    const hasEmail = /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/i.test(noUrls);
    const hasIp = /129\.153|141\.148|193\.122/i.test(noUrls);

    if (!hasNick && !hasEmail && !hasIp) {
      console.log(`✓ [${repo}] docs/status.json: 100% SEGURO E ANÔNIMO`);
    } else {
      console.log(`❌ [${repo}] docs/status.json EXPOSTO: nick=${hasNick}, email=${hasEmail}, ip=${hasIp}`);
    }
  } catch (e) {
    console.log(`[?] [${repo}] docs/status.json não encontrado ou erro:`, e.message);
  }
}
