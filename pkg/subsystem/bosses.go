package subsystem

import (
	"sync"
	"time"
)

// WorldBossEvent representa um chefe mundial ou portal de evento
type WorldBossEvent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SpawnTime   time.Time `json:"spawnTime"`
	MinLevel    int       `json:"minLevel"`
	PortalID    string    `json:"portalId"`
	IsOpen      bool      `json:"isOpen"`
	AlreadyDone bool      `json:"alreadyDone"`
}

// BossScheduler administra a fila de chefes mundiais e dungeons temporárias
type BossScheduler struct {
	mu     sync.RWMutex
	bosses []WorldBossEvent
}

// NewBossScheduler inicializa o agendador
func NewBossScheduler() *BossScheduler {
	return &BossScheduler{
		bosses: make([]WorldBossEvent, 0),
	}
}

// AddBoss registra um chefe na agenda
func (bs *BossScheduler) AddBoss(b WorldBossEvent) {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.bosses = append(bs.bosses, b)
}

// GetReadyBosses verifica quais chefes estão com portal aberto e disponíveis para entrar
func (bs *BossScheduler) GetReadyBosses(characterLevel int) []WorldBossEvent {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	now := time.Now()
	var ready []WorldBossEvent

	for _, b := range bs.bosses {
		if b.AlreadyDone || characterLevel < b.MinLevel {
			continue
		}
		// Se estiver aberto ou dentro da janela de 5 minutos do spawn
		if b.IsOpen || (now.After(b.SpawnTime) && now.Sub(b.SpawnTime) < 10*time.Minute) {
			ready = append(ready, b)
		}
	}
	return ready
}

// MarkCompleted marca que o bot já participou da raid e pegou o loot
func (bs *BossScheduler) MarkCompleted(bossID string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	for i := range bs.bosses {
		if bs.bosses[i].ID == bossID {
			bs.bosses[i].AlreadyDone = true
			bs.bosses[i].IsOpen = false
			break
		}
	}
}
