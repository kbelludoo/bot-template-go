package engine

import (
	"fmt"
	"sync"
)

// BotManager coordena todas as contas ativas do processo
type BotManager struct {
	mu      sync.RWMutex
	bots    []GameBot
	botByID map[string]GameBot
}

func NewBotManager() *BotManager {
	return &BotManager{
		bots:    make([]GameBot, 0),
		botByID: make(map[string]GameBot),
	}
}

func (m *BotManager) Register(b GameBot) error {
	if b == nil || b.ID() == "" {
		return fmt.Errorf("bot deve ter um ID não vazio")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.botByID[b.ID()]; exists {
		return fmt.Errorf("conta '%s' já registrada", b.ID())
	}
	m.bots = append(m.bots, b)
	m.botByID[b.ID()] = b
	return nil
}

func (m *BotManager) GetAllStatuses() []AccountStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]AccountStatus, 0, len(m.bots))
	for _, b := range m.bots {
		list = append(list, b.GetStatus())
	}
	return list
}

func (m *BotManager) GetBot(id string) (GameBot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.botByID[id]
	return b, ok
}

func (m *BotManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.bots)
}

func (m *BotManager) SupplyTicket(accountID, ticket, wsURL string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.bots) == 0 {
		return fmt.Errorf("nenhum bot ativo no momento")
	}

	if accountID != "" {
		if b, ok := m.botByID[accountID]; ok {
			return b.SupplyTicket(ticket, wsURL)
		}
		return fmt.Errorf("conta '%s' não encontrada", accountID)
	}
	if len(m.bots) != 1 {
		return fmt.Errorf("accountId é obrigatório quando há várias contas")
	}

	// Fallback para primeira conta
	return m.bots[0].SupplyTicket(ticket, wsURL)
}

func (m *BotManager) SupplyTurnstileToken(accountID, token string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.bots) == 0 {
		return fmt.Errorf("nenhum bot ativo no momento")
	}

	if accountID != "" {
		if b, ok := m.botByID[accountID]; ok {
			return b.SupplyTurnstileToken(token)
		}
		return fmt.Errorf("conta '%s' não encontrada", accountID)
	}
	if len(m.bots) != 1 {
		return fmt.Errorf("accountId é obrigatório quando há várias contas")
	}

	// Fallback para primeira conta
	return m.bots[0].SupplyTurnstileToken(token)
}

func (m *BotManager) ExecuteAction(accountID, action string, payload map[string]any) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	b, ok := m.botByID[accountID]
	if !ok {
		return nil, fmt.Errorf("conta '%s' não encontrada", accountID)
	}

	return b.ExecuteAction(action, payload)
}
