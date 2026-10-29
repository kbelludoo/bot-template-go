> [!IMPORTANT]
> ### ⚡ SUBSTITUÍDO OFICIALMENTE PELO TEMPLATE RUST
> Este repositório em Go foi substituído oficialmente pelo **[`bot-template-rust`](https://github.com/kbelludoo/bot-template-rust)**.  
> A versão em Rust atinge **100% de paridade funcional**, opera com **Zero-GC**, reduz o consumo de memória de ~50MB para **< 10MB de RAM** e possui latência sub-milissegundo.  
> 🔗 **Novo Repositório Principal**: [https://github.com/kbelludoo/bot-template-rust](https://github.com/kbelludoo/bot-template-rust)  
> 🌐 **Novo Dashboard Central (GitHub Pages)**: [https://kbelludoo.github.io/bot-template-rust/](https://kbelludoo.github.io/bot-template-rust/)

---

# ⚡ Game Bot Starter Kit & VPS Command Center (Go Native Engine - Legado)

Template padronizado de altíssima performance desenvolvido em **Go (Golang)** para criação rápida de bots autônomos para qualquer jogo de navegador / idle (PokeIdle, PokeIdleWorld, LoL Idle, Baiak, Tibia Idle, Sigma, Leveling Idle, Lux Idle, PokeLegacy, Farm Ricochet, etc.), acompanhado de um ecossistema completo de anti-ban, disjuntor heurístico, gestão de fadiga e automação dos 13 subsistemas de jogos idle.

🌐 **Dashboard Central (GitHub Pages):** [https://kbelludoo.github.io/bot-template-go/](https://kbelludoo.github.io/bot-template-go/)

---

## 🌟 Recursos Principais

- ⚡ **Bot Scaffolder Integrado (`cmd/scaffold`):** Gere o esqueleto 100% completo, compilável e testado para qualquer jogo idle com um único comando (`go run ./cmd/scaffold -game "Novo Jogo"`).
- 🛡️ **Anti-Ban Circuit Breaker (`pkg/engine`):** Disjuntor automático que detecta erros 429, 403 e recusas heurísticas do servidor (`experiência rápida demais`, `save malformado`, `outraAba: true`). Desarma o combate, recua para zona segura e aguarda cooldown de 15 a 30 minutos para prevenir suspensões de 7 dias.
- 💤 **Auditoria de Stamina & Fadiga (`pkg/engine`):** Controlador com suporte a stamina decrescente (42h $\to$ 0h) e fadiga crescente (0% a 100%), histerese anti-thrashing (mínimo 5 a 15 min de descanso) e transição para dummies de treino.
- 🧩 **Preservação de Schema Bruto (`pkg/engine/overlay`):** `StateOverlay` preserva 100% dos campos desconhecidos do servidor (`conquistas`, `skins`, `gemas`, `montaria`), enviando sempre envelopes de progresso íntegros e prevenindo o erro de "save malformado".
- 🌐 **Networking Resiliente (`pkg/network`):**
  - **Multi-IP VPS:** Suporte a IP secundário (`bindIP`) em HTTP e WebSockets usando `net.Dialer{LocalAddr: ...}` para isolamento de contas na mesma VPS.
  - **WebSocket Keepalive:** Heartbeat ping/pong em segundo plano, reconexão automática com backoff exponencial e roteador tipado de mensagens (`client.On("action", ...)`).
  - **Stealth Headers:** Impersonação fiel de navegadores modernos (`sec-ch-ua`, `User-Agent`, cookies de sessão).
- 📦 **Suíte Multi-Protocolo (`pkg/protocol`):**
  - `binary.go`: Leitor e Construtor (`ProtoReader` / `ProtoWriter`) de mensagens Protobuf, Varint, Zigzag e `ParseFatigueState` com zero-alocação.
  - `socketio.go`: Parser e formatador Engine.IO v4 / Socket.IO v4 (`0`, `2`/`3`, `40`, `42<id>`, `43<id>`) e extrator de SHA-1 `DiscoverClientRelease`.
  - `colyseus.go`: Opcodes binários do Colyseus v0.16 (`OpJoinRoom=10` com ACK imediato e `OpRoomData=13` com cabeçalho MsgPack).
  - `trpc.go`: Envelopamento e leitura de lotes tRPC v10 (`?batch=1`).
- ⏱️ **Controladores de Ritmo e Stealth (`pkg/engine/pacing.go` & `pkg/network/signer.go`):**
  - `AdaptiveSpeedTuner`: Auto-ajuste dinâmico de cadência por taxa de sucesso e HTTP 429.
  - `AntiWipeTracker`: Cooldown automático e rebaixamento de área após mortes seguidas.
  - `RankingStealthController`: Desaceleração automática ao entrar no Top 3 do ranking público para evitar denúncias manuais.
  - `RequestSigner` & `ComputeDJB2ShardPath`: Assinatura HMAC-SHA256 por requisição com `Idempotency-Key` UUIDv4, sharding DJB2 de 32 bits e `GenerateDeviceFingerprint`.
- 🧩 **Os 13 Subsistemas Plug-and-Play (`pkg/subsystem`):**
  1. `RouteSolver`: Cálculo matemático da melhor rota de caça por nível, fraquezas elementais, kills/h e bestiário.
  2. `InventoryManager`: Auto-sell inteligente de lixo em NPCs, itens protegidos e controle de capacidade da mochila (>80%).
  3. `DailyManager`: Detecção de virada de dia à meia-noite (UTC) e resgate de streaks de login e recompensas diárias.
  4. `PotionManager`: Otimizador de consumo de poções de HP/Mana sem desperdício de cura e auto-recompra.
  5. `PromotionManager`: Aquisição automática de promoções de classe em marcos de nível e manutenção de 100% de bênçãos (Bless 1 a 5).
  6. `PartyCoordinator`: Coordenação de frota multi-contas na mesma VPS com cálculo da taxa coletiva proporcional.
  7. `BossScheduler`: Agendador de spawn de chefes mundiais e dungeons temporárias.
- 🪶 **Ultra-Leve & Zero CGO:** Consumo típico entre **2 e 5 MB de RAM** por processo, compilação estática (`CGO_ENABLED=0`) compatível com qualquer distribuição Linux.
- 📺 **Console Ticker no Terminal:** Estatísticas em tempo real de CPU/RAM, Kills, Ouro e Contas a cada 4 segundos.

---

## 📁 Estrutura do Projeto

```text
bot-template-go/
├── main.go                     # Entrypoint universal (carrega JSON, inicia bots, web server e console ticker)
├── accounts.example.json       # Exemplo limpo para o usuário
├── accounts.json               # Credenciais reais (ignorado pelo git)
├── AGENTS.md                   # 🛡️ Diretrizes universais de IA e anti-ban
├── AI_INSTRUCTIONS.md          # 🤖 Manual de engenharia para desenvolvimento de novos bots
├── go.mod / go.sum             # Módulo Go nativo (dependência única: gorilla/websocket)
│
├── cmd/
│   ├── scaffold/               # ⚡ GERADOR AUTOMÁTICO DE NOVOS BOTS PARA A IA
│   ├── build/                  # Cross-compile para Linux/Windows
│   ├── watchdog/               # Monitor de saúde + auto-cura + alertas Telegram
│   ├── sync-vps-hub/           # Agregador de telemetria da VPS (hub multi-bots)
│   ├── audit/                  # Auditoria de integridade e vazamentos
│   └── telegram-agent/         # Agente de alertas no Telegram
│
├── pkg/
│   ├── engine/                 # Motor: GameBot, CircuitBreaker, StaminaController, StateOverlay, Supervisor
│   ├── network/                # HTTPClient (Multi-IP, Cookies, Retries) e SocketClient (Keepalive, On Router)
│   ├── protocol/               # Binary Protobuf Wire & Varint Reader (Zero-Alloc)
│   ├── subsystem/              # Os 13 Subsistemas: Solver, Inventory, Dailies, Potions, Promotion, Party, Bosses
│   ├── logger/                 # Logging estruturado nativo com log/slog e tags por conta
│   ├── server/                 # Servidor HTTP e Web Dashboard Cyberpunk/Dark embutido (//go:embed)
│   ├── auth/                   # CDP bridge para desafios Cloudflare / Turnstile
│   └── game/                   # Implementação de referência completa (PokeTrue / PokeIdle)
```

---

## 🚀 Como Criar um Bot para um Novo Jogo

### 1. Usando o Bot Scaffolder (Recomendado)
Execute no terminal:
```bash
go run ./cmd/scaffold -game "Meu Jogo Idle" -protocol "websocket" -pkg "meujogo"
```
Isso gera instantaneamente a pasta `pkg/meujogo/` com:
- `types.go`: Configurações de conta e modelo de dados.
- `catalog.go` e `data/areas.json`: Catálogo de áreas e monstros embutido.
- `api.go`: Cliente HTTP/WS com Circuit Breaker e Multi-IP injetados.
- `bot.go`: Implementação de `engine.GameBot` com todos os 13 subsistemas conectados.
- `bot_test.go`: Testes unitários prontos e compiláveis.

### 2. Configure suas contas
Copie `accounts.example.json` para `accounts.json` e adicione seus logins:
```json
[
  {
    "id": "acc-1",
    "name": "Personagem 1",
    "enabled": true,
    "credentials": {
      "token": "seu-token-aqui"
    },
    "settings": {
      "huntTarget": "auto",
      "autoHunt": true,
      "autoSell": true,
      "autoDaily": true
    }
  }
]
```

### 3. Execute localmente
```bash
go run .
```

### 4. Acesse o Dashboard Local
Abra `http://localhost:3001` no navegador para acompanhar o status e executar comandos manuais (curar, mudar área, vender inventário).

---

## 📦 Como Compilar para VPS Linux

Cross-compile nativo (sem CGO, compatível com qualquer distro Linux x86_64):
```bash
# Via script nativo:
go run ./cmd/build linux

# Ou manualmente via terminal:
cmd /c "set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -ldflags=\"-s -w\" -o bot-app-linux ."
```

Envie para a VPS:
```bash
scp bot-app-linux ubuntu@seu-servidor-ip:/home/ubuntu/meu-bot/
```

---

## 🧪 Testes Automatizados

O template inclui testes unitários para todos os módulos:
```bash
go test -v ./...
```
Todos os módulos (`pkg/engine`, `pkg/network`, `pkg/protocol`, `pkg/subsystem`, `pkg/logger`, `pkg/game`, `pkg/example`) passam com 100% de sucesso.
