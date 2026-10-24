package subsystem

import (
	"sync"
	"time"
)

// DailyState armazena o histórico do dia atual
type DailyState struct {
	LastClaimDate string `json:"lastClaimDate"` // YYYY-MM-DD
	StreakDays    int    `json:"streakDays"`
	ClaimedToday  bool   `json:"claimedToday"`
	RerollsUsed   int    `json:"rerollsUsed"`
}

// DailyManager gerencia o ciclo diário de recompensas (reset às 00:00 UTC ou horário do servidor)
type DailyManager struct {
	mu            sync.RWMutex
	lastResetDate string
	state         DailyState
	timezone      *time.Location
}

// NewDailyManager inicializa o gerenciador com fuso horário (ex: UTC ou America/Sao_Paulo)
func NewDailyManager(loc *time.Location) *DailyManager {
	if loc == nil {
		loc = time.UTC
	}
	return &DailyManager{
		timezone: loc,
	}
}

// CheckDailyReset verifica se virou o dia e reinicializa os flags de resgate
func (dm *DailyManager) CheckDailyReset() bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	today := time.Now().In(dm.timezone).Format("2006-01-02")
	if dm.lastResetDate != today {
		dm.lastResetDate = today
		dm.state.ClaimedToday = false
		dm.state.RerollsUsed = 0
		return true // Virou o dia!
	}
	return false
}

// CanClaimRewards verifica se ainda há recompensas diárias pendentes de resgate hoje
func (dm *DailyManager) CanClaimRewards() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	today := time.Now().In(dm.timezone).Format("2006-01-02")
	if dm.state.LastClaimDate != today || !dm.state.ClaimedToday {
		return true
	}
	return false
}

// MarkClaimed registra que o resgate diário foi concluído com sucesso
func (dm *DailyManager) MarkClaimed() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	today := time.Now().In(dm.timezone).Format("2006-01-02")
	dm.state.LastClaimDate = today
	dm.state.ClaimedToday = true
	dm.state.StreakDays++
}

// GetStreak retorna os dias consecutivos de login
func (dm *DailyManager) GetStreak() int {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.state.StreakDays
}
