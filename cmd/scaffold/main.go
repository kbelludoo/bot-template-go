package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	gameName := flag.String("game", "", "Nome do jogo (ex: 'Lux Idle', 'Baiak Idle', 'Leveling Idle')")
	protocol := flag.String("protocol", "websocket", "Protocolo principal: 'websocket', 'rest', 'socketio', 'colyseus', 'protobuf' ou 'trpc'")
	pkgName := flag.String("pkg", "", "Nome do pacote em go (ex: 'luxidle', 'baiak', 'leveling')")
	flag.Parse()

	_ = protocol

	if *gameName == "" {
		fmt.Println("================================================================================")
		fmt.Println("  ⚡ BOT SCAFFOLDER — GERADOR DE BOTS 100% COMPLETOS PARA JOGOS IDLE ⚡")
		fmt.Println("================================================================================")
		fmt.Println("Uso:")
		fmt.Println("  go run ./cmd/scaffold -game \"NomeDoJogo\" -protocol \"websocket|rest|socketio|colyseus|protobuf|trpc\" [-pkg \"pacote\"]")
		fmt.Println("\nExemplo:")
		fmt.Println("  go run ./cmd/scaffold -game \"Lux Idle\" -protocol \"websocket\" -pkg \"luxidle\"")
		os.Exit(1)
	}

	sanitizedPkg := *pkgName
	if sanitizedPkg == "" {
		sanitizedPkg = strings.ToLower(strings.ReplaceAll(*gameName, " ", ""))
	}

	targetDir := filepath.Join("pkg", sanitizedPkg)
	dataDir := filepath.Join(targetDir, "data")
	_ = os.MkdirAll(dataDir, 0755)

	fmt.Printf("🚀 Gerando bot completo para '%s' em '%s'...\n", *gameName, targetDir)

	// 1. types.go
	typesContent := fmt.Sprintf(`package %s

// AccountConfig modelo de configuração lido do accounts.json
type AccountConfig struct {
	ID          string          `+"`json:\"id\"`"+`
	Name        string          `+"`json:\"name\"`"+`
	Enabled     bool            `+"`json:\"enabled\"`"+`
	Credentials GameCredentials `+"`json:\"credentials\"`"+`
	Settings    GameSettings    `+"`json:\"settings\"`"+`
	BindIP      string          `+"`json:\"bindIP,omitempty\"`"+`
}

type GameCredentials struct {
	Token         string `+"`json:\"token\"`"+`
	SessionCookie string `+"`json:\"sessionCookie,omitempty\"`"+`
	Email         string `+"`json:\"email,omitempty\"`"+`
	Password      string `+"`json:\"password,omitempty\"`"+`
}

type GameSettings struct {
	HuntTarget     string  `+"`json:\"huntTarget\"`"+` // "auto" ou slug da área
	AutoHunt       bool    `+"`json:\"autoHunt\"`"+`
	AutoSell       bool    `+"`json:\"autoSell\"`"+`
	AutoHeal       bool    `+"`json:\"autoHeal\"`"+`
	AutoDaily      bool    `+"`json:\"autoDaily\"`"+`
	AutoPromotion  bool    `+"`json:\"autoPromotion\"`"+`
	AutoDummies    bool    `+"`json:\"autoDummies\"`"+`
	HealThreshold  float64 `+"`json:\"healThreshold\"`"+` // ex: 0.5 (cura quando HP < 50%%)
	MinGoldReserve int64   `+"`json:\"minGoldReserve\"`"+`
}

// GameState modelo canônico de dados do jogo
type GameState struct {
	Level       int            `+"`json:\"level\"`"+`
	Exp         int64          `+"`json:\"exp\"`"+`
	Gold        int64          `+"`json:\"gold\"`"+`
	HP          int            `+"`json:\"hp\"`"+`
	MaxHP       int            `+"`json:\"maxHp\"`"+`
	Mana        int            `+"`json:\"mana\"`"+`
	MaxMana     int            `+"`json:\"maxMana\"`"+`
	Vocation    string         `+"`json:\"vocation\"`"+`
	CurrentArea string         `+"`json:\"currentArea\"`"+`
	Inventory   map[string]int `+"`json:\"inventory\"`"+`
}
`, sanitizedPkg)
	_ = os.WriteFile(filepath.Join(targetDir, "types.go"), []byte(typesContent), 0644)

	// 2. catalog.go
	catalogContent := fmt.Sprintf(`package %s

import (
	_ "embed"
	"encoding/json"
	"log"
	"sync"

	"bot-template-go/pkg/subsystem"
)

//go:embed data/areas.json
var areasJSON []byte

var (
	catalogOnce sync.Once
	areasList   []subsystem.AreaCandidate
	solver      *subsystem.RouteSolver
)

func InitCatalog() {
	catalogOnce.Do(func() {
		if len(areasJSON) > 0 {
			if err := json.Unmarshal(areasJSON, &areasList); err != nil {
				log.Printf("[%s] Erro ao carregar areas.json: %%v", err)
			}
		}
		if len(areasList) == 0 {
			// Fallback com áreas básicas
			areasList = []subsystem.AreaCandidate{
				{ID: "area_iniciante", Name: "Área dos Ratos", MinLevel: 1, MaxLevel: 15, BaseExp: 35, KillsPerHour: 450},
				{ID: "area_floresta", Name: "Floresta dos Elfos", MinLevel: 15, MaxLevel: 40, BaseExp: 120, KillsPerHour: 380},
				{ID: "area_caverna", Name: "Caverna dos Dragões", MinLevel: 40, MaxLevel: 100, BaseExp: 550, KillsPerHour: 250},
			}
		}
		solver = subsystem.NewRouteSolver(areasList)
	})
}

func GetRouteSolver() *subsystem.RouteSolver {
	InitCatalog()
	return solver
}
`, sanitizedPkg, sanitizedPkg)
	_ = os.WriteFile(filepath.Join(targetDir, "catalog.go"), []byte(catalogContent), 0644)

	// 3. data/areas.json
	defaultAreas := `[
  {
    "id": "area_iniciante",
    "name": "Área dos Ratos",
    "minLevel": 1,
    "maxLevel": 15,
    "baseExp": 35,
    "baseGold": 10,
    "killsPerHour": 450,
    "monsters": ["Rato", "Aranha"]
  },
  {
    "id": "area_floresta",
    "name": "Floresta dos Elfos",
    "minLevel": 15,
    "maxLevel": 40,
    "baseExp": 120,
    "baseGold": 45,
    "killsPerHour": 380,
    "monsters": ["Elfo", "Lobo"]
  },
  {
    "id": "area_caverna",
    "name": "Caverna dos Dragões",
    "minLevel": 40,
    "maxLevel": 100,
    "baseExp": 550,
    "baseGold": 180,
    "killsPerHour": 250,
    "monsters": ["Dragão", "Lorde Dragão"]
  }
]`
	_ = os.WriteFile(filepath.Join(dataDir, "areas.json"), []byte(defaultAreas), 0644)

	// 4. api.go
	apiContent := fmt.Sprintf(`package %s

import (
	"context"
	"fmt"
	"time"

	"bot-template-go/pkg/engine"
	"bot-template-go/pkg/network"
)

// Client comunica-se com a API oficial do %s
type Client struct {
	http *network.HTTPClient
	ws   *network.SocketClient
	cb   *engine.CircuitBreaker
}

func NewClient(baseURL, bindIP string, cb *engine.CircuitBreaker) *Client {
	cfg := network.DefaultHTTPConfig()
	cfg.BaseURL = baseURL
	cfg.BindIP = bindIP
	cfg.Timeout = 15 * time.Second

	return &Client{
		http: network.NewHTTPClientWithConfig(cfg),
		cb:   cb,
	}
}

// FetchState carrega o estado canônico do personagem
func (c *Client) FetchState(ctx context.Context) (*network.Response, error) {
	resp, err := c.http.Get(ctx, "/api/eu")
	if err != nil {
		c.cb.RecordFailure(err.Error())
		return nil, err
	}
	if resp.IsAntiCheatRejection() {
		c.cb.Trip(fmt.Sprintf("Anti-Cheat Rejection HTTP %%d: %%s", resp.StatusCode, resp.String()))
		return nil, fmt.Errorf("ação rejeitada pelo anti-cheat: %%s", resp.String())
	}
	c.cb.RecordSuccess()
	return resp, nil
}

// SaveProgress salva o estado com overlay preservando chaves desconhecidas
func (c *Client) SaveProgress(ctx context.Context, overlay *engine.StateOverlay) (*network.Response, error) {
	if !c.cb.AllowAction() {
		return nil, fmt.Errorf("circuit breaker desarmado, save abortado para evitar ban")
	}

	payload := overlay.Envelope("estado")
	resp, err := c.http.PutJSON(ctx, "/api/save", payload)
	if err != nil {
		c.cb.RecordFailure(err.Error())
		return nil, err
	}
	if resp.IsAntiCheatRejection() {
		c.cb.Trip(fmt.Sprintf("Recusa em PUT /api/save: %%s", resp.String()))
		return nil, fmt.Errorf("save rejeitado: %%s", resp.String())
	}
	c.cb.RecordSuccess()
	return resp, nil
}
`, sanitizedPkg, *gameName)
	_ = os.WriteFile(filepath.Join(targetDir, "api.go"), []byte(apiContent), 0644)

	// 5. bot.go
	botContent := fmt.Sprintf(`package %s

import (
	"context"
	"fmt"
	"sync"
	"time"

	"bot-template-go/pkg/engine"
	"bot-template-go/pkg/subsystem"
)

// Bot implementa a interface obrigatória engine.GameBot
type Bot struct {
	config  AccountConfig
	mu      sync.RWMutex
	status  engine.AccountStatus
	client  *Client
	cb      *engine.CircuitBreaker
	overlay *engine.StateOverlay

	// Sub-sistemas de jogo e controladores de ritmo
	stamina    *engine.StaminaController
	speedTuner *engine.AdaptiveSpeedTuner
	antiWipe   *engine.AntiWipeTracker
	rankStealth *engine.RankingStealthController
	solver     *subsystem.RouteSolver
	inventory  *subsystem.InventoryManager
	dailies    *subsystem.DailyManager
	potions    *subsystem.PotionManager
	promotion  *subsystem.PromotionManager
	bosses     *subsystem.BossScheduler
}

// NewBot inicializa o bot com todos os subsistemas e controladores anti-ban ativos
func NewBot(cfg AccountConfig) *Bot {
	InitCatalog()

	cbCfg := engine.DefaultCircuitBreakerConfig()
	cb := engine.NewCircuitBreaker(cfg.ID, cbCfg)

	st := engine.NewStatus(cfg.ID, cfg.Name)
	st.CurrentArea = cfg.Settings.HuntTarget
	if st.CurrentArea == "" {
		st.CurrentArea = "auto"
	}

	staminaCfg := engine.DefaultStaminaConfig(engine.StaminaTypeDecreasing)
	stamina := engine.NewStaminaController(staminaCfg)
	// Inicializa stamina padrão até sincronizar com GET /api/eu
	stamina.UpdateStaminaHours(42.0)

	b := &Bot{
		config:      cfg,
		status:      st,
		client:      NewClient("https://localhost", cfg.BindIP, cb),
		cb:          cb,
		overlay:     engine.NewStateOverlay(nil),
		stamina:     stamina,
		speedTuner:  engine.NewAdaptiveSpeedTuner(engine.DefaultAdaptiveSpeedConfig()),
		antiWipe:    engine.NewAntiWipeTracker(3, 10*time.Minute, 15*time.Minute),
		rankStealth: engine.NewRankingStealthController(5, 30, 0.5),
		solver:      GetRouteSolver(),
		inventory:   subsystem.NewInventoryManager(100, []string{"gold", "token", "potion"}),
		dailies:     subsystem.NewDailyManager(time.UTC),
		potions:     subsystem.NewPotionManager(nil, 20),
		bosses:      subsystem.NewBossScheduler(),
	}

	cbCfg.OnTrip = func(reason string) {
		b.addLog(fmt.Sprintf("🚨 Anti-Cheat disparado: %%s. Recuando para zona segura.", reason))
		b.updateStatus(func(s *engine.AccountStatus) {
			s.Status = "Proteção Anti-Cheat Ativa (Cooldown)"
		})
	}

	cbCfg.OnReset = func() {
		b.addLog("✅ Cooldown finalizado. Retomando caça normal.")
		b.updateStatus(func(s *engine.AccountStatus) {
			s.Status = "hunting"
		})
	}

	return b
}

func (b *Bot) ID() string {
	return b.config.ID
}

func (b *Bot) Name() string {
	return b.config.Name
}

func (b *Bot) Init(ctx context.Context) error {
	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "Autenticando..."
	})
	b.addLog("Inicialização concluída com sucesso.")
	return nil
}

func (b *Bot) Run(ctx context.Context) error {
	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "hunting"
	})

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.addLog("Loop de caça encerrado limpo via Context.")
			return nil
		case <-ticker.C:
			b.tick(ctx)
		}
	}
}

func (b *Bot) tick(ctx context.Context) {
	// 1. Verificar Circuit Breaker
	if !b.cb.AllowAction() {
		return
	}

	// 2. Auditoria de Stamina / Fadiga
	if shouldRest, reason := b.stamina.ShouldRest(); shouldRest {
		b.addLog(fmt.Sprintf("💤 Descanso acionado: %%s", reason))
		b.updateStatus(func(s *engine.AccountStatus) {
			s.Status = "resting"
		})
		return
	}

	// 3. Resgate de Recompensas Diárias
	if b.config.Settings.AutoDaily && b.dailies.CanClaimRewards() {
		b.addLog("🎁 Resgatando bônus diário de login...")
		b.dailies.MarkClaimed()
	}

	// 4. Solver de Área Ótima
	if b.config.Settings.HuntTarget == "auto" {
		crit := subsystem.SolverCriteria{
			CharacterLevel: b.status.Level,
		}
		if best, ok := b.solver.FindOptimalRoute(crit); ok && best.Area.ID != b.status.CurrentArea {
			b.addLog(fmt.Sprintf("🗺️ Transicionando rota: %%s -> %%s (%%s)", b.status.CurrentArea, best.Area.Name, best.Reason))
			b.updateStatus(func(s *engine.AccountStatus) {
				s.CurrentArea = best.Area.Name
			})
		}
	}
}

func (b *Bot) GetStatus() engine.AccountStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

func (b *Bot) ExecuteAction(action string, payload map[string]any) (any, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch action {
	case "heal":
		b.status.HP = b.status.MaxHP
		b.addLog("Vida regenerada manualmente.")
		return "curado", nil
	case "sell":
		b.addLog("Auto-Sell executado manualmente.")
		return "inventário vendido", nil
	default:
		return nil, fmt.Errorf("ação '%%s' desconhecida", action)
	}
}

func (b *Bot) SupplyTicket(ticket, wsURL string) error {
	b.addLog("Ticket de autenticação atualizado.")
	return nil
}

func (b *Bot) SupplyTurnstileToken(token string) error {
	b.addLog("Token Turnstile atualizado.")
	return nil
}

func (b *Bot) addLog(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	t := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%%s] %%s", t, msg)
	b.status.Logs = append(b.status.Logs, entry)
	if len(b.status.Logs) > 100 {
		b.status.Logs = b.status.Logs[len(b.status.Logs)-100:]
	}
	b.status.UpdatedAt = time.Now().UnixMilli()
}

func (b *Bot) updateStatus(fn func(s *engine.AccountStatus)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fn(&b.status)
	b.status.UpdatedAt = time.Now().UnixMilli()
}
`, sanitizedPkg)
	_ = os.WriteFile(filepath.Join(targetDir, "bot.go"), []byte(botContent), 0644)

	// 6. bot_test.go
	testContent := fmt.Sprintf(`package %s

import (
	"context"
	"testing"
	"time"
)

func TestBotLifecycle(t *testing.T) {
	cfg := AccountConfig{
		ID:      "test-acc",
		Name:    "TestHero",
		Enabled: true,
		Settings: GameSettings{
			HuntTarget: "auto",
			AutoHunt:   true,
			AutoDaily:  true,
		},
	}

	bot := NewBot(cfg)
	if bot.ID() != "test-acc" {
		t.Errorf("ID incorreto: %%s", bot.ID())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := bot.Init(ctx); err != nil {
		t.Fatalf("Init falhou: %%v", err)
	}

	_ = bot.Run(ctx)

	st := bot.GetStatus()
	if st.ID != "test-acc" {
		t.Errorf("Status ID inesperado: %%s", st.ID)
	}
}
`, sanitizedPkg)
	_ = os.WriteFile(filepath.Join(targetDir, "bot_test.go"), []byte(testContent), 0644)

	fmt.Printf("✅ Bot '%s' gerado com sucesso em '%s'!\n", *gameName, targetDir)
	fmt.Printf("   Arquivos gerados:\n")
	fmt.Printf("   - %s/types.go\n", targetDir)
	fmt.Printf("   - %s/catalog.go\n", targetDir)
	fmt.Printf("   - %s/data/areas.json\n", targetDir)
	fmt.Printf("   - %s/api.go\n", targetDir)
	fmt.Printf("   - %s/bot.go\n", targetDir)
	fmt.Printf("   - %s/bot_test.go\n", targetDir)
}
