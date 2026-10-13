const { execSync } = require('child_process');

const repos = [
  'leveling-bot',
  'pokeidle-go',
  'pokeidleworld-bot',
  'sigma-bot',
  'tibidle-bot',
  'bot-template-go',
  'huntera-bot',
  'kick-miner-go',
  'twitch-miner-go'
];

for (const r of repos) {
  try {
    const rawB64 = execSync(`gh api repos/kbelludoo/${r}/contents/docs/index.html --jq .content 2>/dev/null`, { timeout: 10000 }).toString().trim();
    if (!rawB64) {
      console.log(`[-] ${r}/docs/index.html: não encontrado`);
      continue;
    }
    const raw = Buffer.from(rawB64, 'base64').toString('utf8');
    const matches = [];
    raw.split('\n').forEach((line, idx) => {
      if (/(kbelludoo|lohmoreira|\blrn\b)/i.test(line) && !line.includes('github.com') && !line.includes('github.io')) {
        matches.push(`L${idx+1}: ${line.trim()}`);
      }
    });
    if (matches.length > 0) {
      console.log(`[!] ${r}/docs/index.html TEM ${matches.length} OCORRÊNCIAS:`);
      matches.forEach(m => console.log('   ' + m));
    } else {
      console.log(`[✓] ${r}/docs/index.html 100% limpo`);
    }
  } catch (e) {
    console.log(`[-] ${r}: ${e.message}`);
  }
}
