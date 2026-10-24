# 🤖 AI_INSTRUCTIONS.md — MANUAL DE ENGENHARIA PARA IAs

> **DESTINATÁRIO:** Qualquer modelo ou agente de IA (Antigravity, Claude, ChatGPT, etc.) encarregado de criar ou modificar um bot de jogo em Go.
> **LEIA ESTE ARQUIVO ANTES DE TOCAR EM QUALQUER LINHA DE CÓDIGO.**

---

## 🎯 OBJETIVO PRINCIPAL
Você está construindo ou aprimorando bots autônomos nativos em **Go (Golang)** de altíssima performance e baixíssimo consumo de recursos para jogos de navegador/idle (ex: PokeIdle, PokeIdleWorld, LoL Idle, Baiak, Tibia Idle, Sigma, Leveling Idle, Lux Idle, PokeLegacy, Farm Ricochet, etc.).

**Invariantes Inegociáveis:**
1. **Cada Jogo é um Binário Independente:** Nunca misture múltiplos jogos no mesmo executável. Se o jogo X cair ou mudar de protocolo, os outros jogos não podem ser afetados.
2. **Consumo de Memória Cirúrgico:** O processo deve consumir **menos de 15 MB de RAM** (geralmente entre 2 e 5 MB). Nunca use browsers headless (Puppeteer/Playwright/Chromium) na VPS para rodar o loop de jogo. Conecte diretamente via WebSockets ou HTTP.
3. **Resiliência Total contra Crash:** Toda conta roda em Goroutine protegida com `engine.RunWithSupervisor` com `recover()` e cancelamento limpo via `context.Context`. Um panic de unmarshal ou erro de rede NUNCA deve derrubar o executável principal.
4. **Isolamento de Portas Web:** Cada bot tem sua própria porta no dashboard embutido (ex: Huntera usa 3000, Tibidle usa 3001, Sigma usa 3002). Nunca crie conflito de portas.
5. **Segurança Absoluta de Segredos (Zero Hardcoding):** NUNCA insira tokens de API, senhas, cookies, chaves SSH (`*.key`), certificados (`*.pem`), tokens (`*.tok`) ou IPs públicos da VPS diretamente em arquivos commitáveis. Credenciais ficam no `accounts.json` (no `.gitignore`) ou variáveis de ambiente.
6. **Modo Stealth & Anti-Ban (Zero Nicks Públicos):** NUNCA exponha o nome real do personagem nos dashboards públicos ou `status.json`. Exiba apenas vocação/classe, índice anônimo (`Conta #1`) e métricas técnicas.
7. **Proteção Anti-Ban Circuit Breaker Obrigatória:** Todo bot DEVE integrar `pkg/engine.CircuitBreaker`. Diante de qualquer recusa de save, rate-limit ou erro de anti-cheat (`experiência rápida demais`, `recusado: true`), o bot NUNCA deve insistir em novos saves. Deve desarmar o combate, recuar para zona segura e aguardar cooldown de 15 a 30 minutos.
8. **Proibição Absoluta de Bots "Barebones" (Completude Máxima):** É proibido construir bots simplistas que apenas batem em monstros e coletam ouro. O bot DEVE implementar os 13 subsistemas do jogo: Presas, Recompensas Diárias, Missões/Tasks, Bênçãos, Chefes Mundiais, Forja/Gemas, Promoção de Vocação, Dummies de Treino e Gestão de Inventário.

---

## 📁 ESTRUTURA DO TEMPLATE E MÓDULOS NATIVOS

```text
bot-template-go/
├── main.go                     # Entrypoint universal (carrega JSON, inicia bots, web server e console ticker)
├── accounts.json               # Credenciais reais (IGNORADO PELO GIT)
├── accounts.example.json       # Exemplo limpo para o usuário
├── AGENTS.md                   # Diretrizes universais de IA e anti-ban
├── AI_INSTRUCTIONS.md          # Manual de engenharia para desenvolvimento de novos bots
├── go.mod / go.sum             # Módulo Go nativo (dependência única: gorilla/websocket)
│
├── cmd/
│   ├── scaffold/               # ⚡ GERADOR AUTOMÁTICO DE NOVOS BOTS PARA A IA ⚡
│   │   └── main.go             # Executa: go run ./cmd/scaffold -game "Jogo" -pkg "slug"
│   ├── watchdog/               # Monitor de integridade e reinício de processos
│   ├── telegram-agent/         # Bot de alertas e telemetria remota via Telegram
│   └── vps-agent/              # Agente de saúde e métricas de hardware da VPS
│
├── pkg/
│   ├── engine/                 # [NÚCLEO DE RESILIÊNCIA E ANTI-BAN]
│   │   ├── bot.go              # Interface GameBot e struct AccountStatus
│   │   ├── circuit_breaker.go  # Disjuntor anti-ban, Half-Open (BeginRecovery/FinishRecovery) e taxa coletiva em party
│   │   ├── overlay.go          # ⭐ StateOverlay (RawEstado): preservação de schema, ContractEnvelope e ReconcileStaleSave
│   │   ├── stamina.go          # ⭐ StaminaController: auditoria de fadiga/stamina com guardas de inicialização
│   │   ├── pacing.go           # ⭐ AdaptiveSpeedTuner, AntiWipeTracker, RankingStealthController e CircularPatrolPosition
│   │   ├── manager.go          # Gerenciador concorrente (thread-safe) de contas ativas com validação de IDs
│   │   └── supervisor.go       # Goroutine runner com backoff exponencial, estabilidade de 1 min e context cancellation
│   │
│   ├── network/                # [NETWORKING RESILIENTE DE ALTA PERFORMANCE]
│   │   ├── http_client.go      # Cliente HTTP com Multi-IP (bindIP), fila MaxConcurrent, sharding dinâmico e retries
│   │   ├── signer.go           # ⭐ RequestSigner (HMAC-SHA256 + Idempotency-Key), DJB2 Sharding e Device Fingerprint
│   │   └── socket.go           # Cliente WebSocket com keepalive ping/pong, Multi-IP, roteador On e binário
│   │
│   ├── protocol/               # [PARSERS E CONSTRUTORES DE PROTOCOLOS]
│   │   ├── binary.go           # ProtoReader e ProtoWriter sem schema (.proto) + ParseFatigueState
│   │   ├── socketio.go         # Engine.IO v4 / Socket.IO v4 (0, 2/3, 40, 42<ack>, 43<ack>) + DiscoverClientRelease
│   │   ├── colyseus.go         # Colyseus v0.16 binário (OpJoinRoom=10 ACK, OpRoomData=13 msgpack framing)
│   │   └── trpc.go             # Envelopamento e desenvelopamento tRPC v10 Batch (?batch=1)
│   │
│   ├── auth/                   # [BYPASS DE TURNSTILE E CAMUFLAGEM CDP]
│   │   ├── cdp.go              # Extrator de sessão/cookies via Chrome DevTools Protocol
│   │   └── stealth_hook.go     # StealthInitScript, TurnstileInterceptScript (DummyWebSocket) e BuildCleanAntiBotLaudo
│   │
│   ├── subsystem/              # [OS SUBSISTEMAS PLUG-AND-PLAY]
│   │   ├── solver.go           # RouteSolver: benchmark empírico de EXP/h, SimulateCombat (TTK/Sustain) e matriz Pokémon 18x18
│   │   ├── inventory.go        # InventoryManager: auto-sell de lixo no NPC, protected items e monitor de peso
│   │   ├── dailies.go          # DailyManager: detector de reset à meia-noite e resgate de streaks
│   │   ├── potions.go          # PotionManager: cálculo da melhor poção sem desperdício e auto-restock
│   │   ├── promotion.go        # PromotionManager: compra de classe em marcos de nível e bênçãos (Bless 1-5)
│   │   ├── party.go            # PartyCoordinator: sincronia pelo menor nível, bestiário compartilhado e LockHunt
│   │   ├── bosses.go           # BossScheduler: monitoramento de portais e chefes mundiais
│   │   └── lurker.go           # Stream Lurker: payloads de presença e resgate de drops para Twitch GQL/IRC e Kick WS
│   │
│   ├── logger/                 # [LOGGING ESTRUTURADO slog]
│   │   └── logger.go           # Logger limpo com tags coloridas por conta e zero overhead
│   │
│   ├── server/                 # [DASHBOARD EMBUTIDO]
│   │   ├── server.go           # Servidor HTTP para /api/status, /api/action, /api/ticket com shutdown gracioso
│   │   └── index.html          # Web Dashboard Dark embutido via //go:embed
│   │
│   └── game/                   # Implementação de referência (PokeTrue / PokeIdle)
│       ├── types.go            # Modelos de pacotes e estado
│       ├── catalog.go          # Catálogo de dados (hunts, dex, loot) com //go:embed
│       ├── api.go              # Cliente HTTP/WS da API
│       └── bot.go              # Implementação de engine.GameBot
```

---

## 🚀 PASSO A PASSO PARA A IA CRIAR UM NOVO BOT EM MINUTOS

Sempre que o usuário pedir para criar um bot para um jogo novo, a IA DEVE seguir este roteiro:

### Passo 1: Executar o Scaffolder do Template
Não crie arquivos manualmente do zero. Use o comando nativo:
```bash
go run ./cmd/scaffold -game "Nome do Jogo" -protocol "websocket" -pkg "nomedojogo"
```
Isso gera instantaneamente em `pkg/<nomedojogo>/`:
- `types.go`: Configurações de conta e modelo de pacotes.
- `catalog.go`: Catálogo com `//go:embed data/areas.json`.
- `data/areas.json`: Arquivo inicial de áreas de caça.
- `api.go`: Cliente de comunicação com Circuit Breaker e Multi-IP injetados.
- `bot.go`: Estrutura completa de `engine.GameBot` com todos os 13 subsistemas pré-conectados.
- `bot_test.go`: Teste unitário compilável e executável.

### Passo 2: Reconhecimento e Engenharia Reversa (DevTools)
1. **Network (WS vs REST):**
   - Inspecione se os pacotes trafegam por WebSocket (`wss://`) ou REST Polling (`GET/PUT`).
   - Se for binário (frames ilegíveis na aba WS), use `pkg/protocol.NewProtoReader` para decodificar varints e tags.
2. **Autenticação:**
   - Como o jogo autentica? Header `Authorization: Bearer <token>`, cookie de sessão ou pacote de login `{t: "login"}`?
3. **Auditoria de Fadiga / Stamina (CRÍTICO):**
   - O jogo tem limitador de caça? Configure `engine.NewStaminaController(cfg)` no bot:
     - `StaminaTypeIncreasingFatigue`: Fadiga 0% a 100% (parar em $\ge 80\%$, descanso mínimo 5 min).
     - `StaminaTypeDecreasing`: Stamina decrescente (parar ao atingir limite crítico, ex: $\le 14\text{h}$).
4. **Extração dos Dados do Jogo (Mineração JS):**
   - Baixe o bundle client-side (`main.js`, `bundle.js`, etc.) e extraia todas as áreas, monstros, fraquezas e itens para `pkg/<jogo>/data/areas.json` e `pkg/<jogo>/data/items.json`.

### Passo 3: Implementar os Handlers no `bot.go`
Utilize os módulos do `pkg/subsystem/`:
- **Rota Ótima:** Use `b.solver.FindOptimalRoute(crit)` para transicionar automaticamente para a melhor área conforme o personagem sobe de nível.
- **Auto-Sell:** Ao atingir $>80\%$ de inventário, filtre com `b.inventory.GetSellableBatch(stash)` e venda apenas o lixo seguro no NPC.
- **Preservação de Schema (`RawEstado`):** Sempre que salvar progresso, faça overlay com `b.overlay.Set(...)` ou `b.overlay.Merge(...)` e envie `b.overlay.Envelope("estado")` para prevenir erro de "save malformado".
- **Dailies e Streak:** Use `b.dailies.CanClaimRewards()` no tick para resgatar bônus diários.
- **Promoção e Bênçãos:** Use `b.promotion.CheckEligiblePromotion(...)` para comprar rank up de classe e manter Bless 1 a 5 sempre ativas.

### Passo 4: Testar e Validar
Execute o suite de testes:
```bash
go test -v ./pkg/<nomedojogo>/...
go build .
```

---

## 🧠 REGRAS DE OURO DE ENGENHARIA EM GO

1. **Contexto Sempre:** Todo loop com ticker DEVE verificar `case <-ctx.Done(): return nil` para permitir encerramento limpo via SIGTERM.
2. **Escrita Concorrente em WebSocket:** NUNCA chame métodos de escrita no WebSocket sem mutex. O cliente `pkg/network.SocketClient` já gerencia isso nativamente com `writeMu`.
3. **Multi-IP (VPS Anti-Ban):** Ao rodar mais de uma conta no mesmo jogo, configure `"bindIP": "10.0.0.X"` no `accounts.json`. O `pkg/network` faz o bind automaticamente no socket do sistema operacional.
4. **Zero CGO:** Mantenha `CGO_ENABLED=0` para que o binário compile de forma estática e rode em qualquer VPS Linux sem depender de glibc externa.
5. **Memória Abaixo de 15 MB:** Evite alocações excessivas dentro de loops de combate. Reutilize buffers e evite deserializar structs gigantes quando puder ler campos pontuais.
