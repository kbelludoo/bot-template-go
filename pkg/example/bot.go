package example

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
		b.addLog(fmt.Sprintf("🚨 Anti-Cheat disparado: %s. Recuando para zona segura.", reason))
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
		b.addLog(fmt.Sprintf("💤 Descanso acionado: %s", reason))
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
			b.addLog(fmt.Sprintf("🗺️ Transicionando rota: %s -> %s (%s)", b.status.CurrentArea, best.Area.Name, best.Reason))
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
		return nil, fmt.Errorf("ação '%s' desconhecida", action)
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
	entry := fmt.Sprintf("[%s] %s", t, msg)
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
