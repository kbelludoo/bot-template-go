# ⚡ Game Bot Starter Kit & VPS Command Center (Go Native Engine)

Template padronizado de altíssima performance desenvolvido em **Go (Golang)** para criação rápida de bots autônomos para qualquer jogo de navegador / idle, além de um **Command Center Central** no GitHub Pages monitorando todos os bots em execução na VPS em tempo real.

🌐 **Dashboard Central (GitHub Pages):** [https://kbelludoo.github.io/bot-template-go/](https://kbelludoo.github.io/bot-template-go/)

---

## 🌟 Recursos Principais

- 📊 **VPS Command Center Integrado:** Acompanhe a saúde da VPS (CPU load, memória RAM, uptime) e todos os bots rodando simultaneamente (Huntera, Tibidle, Pokemon Sigma) em uma única interface Cyberpunk / Dark.
- 🪶 **Ultra-leve:** Consumo típico entre **2 e 5 MB de RAM** por processo.
- 🛡️ **Supervisão Contínua:** Múltiplas contas em Goroutines isoladas com auto-reconexão e recuperação contra panic via `engine.RunWithSupervisor`.
- 🌐 **Web Dashboard Embutido:** Servidor HTTP local independente (`/api/status` e `/api/action`).
- 📺 **Console Ticker no Terminal:** Estatísticas de CPU/RAM, Kills, Ouro e Contas ativas limpas a cada 4 segundos.
- 🚀 **Pronto para VPS:** Zero navegadores na VPS. Script de compilação cruzada para Linux e template de serviço systemd incluídos.

---

## 📁 Estrutura do Projeto

```text
bot-template-go/
├── AI_INSTRUCTIONS.md          # 🤖 MANUAL COMPLETO PARA QUALQUER IA LER E GERAR NOVOS BOTS
├── accounts.example.json       # Modelo de configuração de contas
├── accounts.json               # Suas credenciais reais (ignorado pelo git)
├── main.go                     # Ponto de entrada universal
├── go.mod / go.sum             # Módulo Go nativo
│
├── docs/                       # 🌐 GITHUB PAGES (COMMAND CENTER AO VIVO)
│   ├── index.html              # Dashboard Cyberpunk agregador
│   └── status.json             # Telemetria em tempo real da VPS e bots
│
├── pkg/
│   ├── engine/                 # Motor: ciclo de vida, supervisor e gerenciador
│   ├── network/                # Conexões: WebSocket thread-safe e cliente HTTP
│   ├── auth/                   # Autenticação e CDP bridge para Cloudflare
│   ├── server/                 # Servidor HTTP e Dashboard embutido local
│   └── game/                   # 🎮 AQUI ENTRA A LÓGICA DO NOVO JOGO
│
└── scripts/
    ├── build-linux.bat         # Compila para Linux na VPS com 1 clique (Windows)
    ├── bot.service             # Template para systemd na VPS Ubuntu
    ├── sync-vps-hub.js         # Agregador de telemetria da VPS para o GitHub Pages
    └── sync-vps-hub.sh         # Script cron de sincronização automática
```

---

## 🚀 Como Criar um Bot para um Novo Jogo

1. **Copie esta pasta:** Duplique a pasta `bot-template-go` e renomeie para o nome do jogo (ex: `meu-novo-bot`).
2. **Abra com a IA:** Peça para a IA ler o arquivo `AI_INSTRUCTIONS.md`. Ela já saberá exatamente a estrutura, os métodos e como codificar a pasta `pkg/game/`.
3. **Configure suas contas:** Copie `accounts.example.json` para `accounts.json` e adicione seus logins.
4. **Execute localmente:**
   ```bash
   go run .
   ```
5. **Acesse o Dashboard Local:** Abra `http://localhost:3001` (ou a porta configurada) no navegador.

---

## 📦 Como Compilar para VPS Linux

No Windows (usando o batch incluído):
```cmd
scripts\build-linux.bat
```
Ou manualmente no terminal:
```cmd
cmd /c "set GOOS=linux&& set GOARCH=amd64&& go build -ldflags=\"-s -w\" -o bot-app-linux ."
```

Envie para a VPS:
```bash
scp bot-app-linux ubuntu@129.153.157.17:/home/ubuntu/meu-bot/
```

---

## 🔄 Sincronização 24/7 com GitHub Pages

Para manter o status da VPS e de todos os bots atualizado automaticamente no GitHub Pages a cada 2 minutos via cron:

```bash
# Na VPS:
chmod +x /home/ubuntu/sync-vps-hub.sh
crontab -e
# Adicione a linha:
*/2 * * * * /home/ubuntu/sync-vps-hub.sh >/dev/null 2>&1
```
