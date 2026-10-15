package game

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"bot-template-go/pkg/engine"
	"bot-template-go/pkg/network"
)

type Bot struct {
	config  AccountConfig
	mu      sync.RWMutex
	status  engine.AccountStatus
	socket  *network.SocketClient
	httpCli *network.HTTPClient

	// Controle interno da sessão do jogo
	isHunting     bool
	lastActionAt  time.Time
	ticketChan    chan string
	turnstileChan chan string
}

func NewBot(cfg AccountConfig) *Bot {
	st := engine.NewStatus(cfg.ID, cfg.Name)
	st.CurrentArea = cfg.Settings.HuntTarget

	return &Bot{
		config:        cfg,
		status:        st,
		httpCli:       network.NewHTTPClient("https://game-api.example.com", 10*time.Second),
		ticketChan:    make(chan string, 1),
		turnstileChan: make(chan string, 1),
	}
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

	// 1. Exemplo de autenticação via HTTP API ou validação de cookies
	log.Printf("[%s] Validando credenciais para %s ...", b.ID(), b.config.Credentials.Email)
	time.Sleep(300 * time.Millisecond) // Simula handshake

	// 2. Prepara conexão de rede
	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "Pronto para conectar"
		s.Level = 1
		s.HP = 100
		s.MaxHP = 100
		s.Mana = 50
		s.MaxMana = 50
	})

	return nil
}

func (b *Bot) Run(ctx context.Context) error {
	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "hunting"
	})
	b.addLog("Conexão com o jogo ativa. Iniciando rotinas de caça.")

	// Loop de simulação/ações periódicas do bot
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			b.performGameStep()
		}
	}
}

func (b *Bot) performGameStep() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Exemplo de atualização de status durante a caçada
	b.status.Kills++
	b.status.Gold += 15
	b.status.ExpPerHour = 85000
	b.status.DurationSec += 2
	b.status.UpdatedAt = time.Now().UnixMilli()

	if b.status.Kills%20 == 0 {
		b.status.Level++
		b.status.MaxHP += 15
		b.status.HP = b.status.MaxHP
		b.status.Logs = append(b.status.Logs, fmt.Sprintf("Parabéns! Nível %d alcançado!", b.status.Level))
		if len(b.status.Logs) > 10 {
			b.status.Logs = b.status.Logs[1:]
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
	case "autoHunt":
		b.config.Settings.AutoHunt = !b.config.Settings.AutoHunt
		if b.config.Settings.AutoHunt {
			b.status.Status = "hunting"
			b.status.Logs = append(b.status.Logs, "Auto-hunt ativado.")
		} else {
			b.status.Status = "idle"
			b.status.Logs = append(b.status.Logs, "Auto-hunt pausado.")
		}
		return map[string]any{"autoHunt": b.config.Settings.AutoHunt}, nil

	case "autoHeal":
		b.status.HP = b.status.MaxHP
		b.status.Mana = b.status.MaxMana
		b.status.Logs = append(b.status.Logs, "Vida e mana restauradas com sucesso.")
		return "Curado com sucesso!", nil

	case "autoSell":
		gained := int64(150)
		b.status.Gold += gained
		b.status.Logs = append(b.status.Logs, fmt.Sprintf("Loot vendido: +%d moedas.", gained))
		return fmt.Sprintf("Vendido! +%d gold", gained), nil

	default:
		return nil, fmt.Errorf("ação desconhecida: %s", action)
	}
}

func (b *Bot) SupplyTicket(ticket, wsURL string) error {
	log.Printf("[%s] Recebido ticket externo de jogo: %s", b.ID(), ticket)
	select {
	case b.ticketChan <- ticket:
	default:
	}
	return nil
}

func (b *Bot) SupplyTurnstileToken(token string) error {
	log.Printf("[%s] Recebido Turnstile Token roubado: %d chars", b.ID(), len(token))
	select {
	case b.turnstileChan <- token:
	default:
	}
	return nil
}

func (b *Bot) updateStatus(fn func(s *engine.AccountStatus)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fn(&b.status)
	b.status.UpdatedAt = time.Now().UnixMilli()
}

func (b *Bot) addLog(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status.Logs = append(b.status.Logs, msg)
	if len(b.status.Logs) > 10 {
		b.status.Logs = b.status.Logs[1:]
	}
}
