package subsystem

import (
	"fmt"
	"math"
	"sync"
	"time"

	"bot-template-go/pkg/engine"
)

// PartyMember representa um membro do grupo cooperativo
type PartyMember struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Role     string    `json:"role"` // "tank", "dps", "healer"
	Level    int       `json:"level"`
	LastSeen time.Time `json:"lastSeen"`
}

// PartyCoordinator coordena múltiplos bots na mesma VPS para compartilhamento seguro de EXP e sincronia de rota
type PartyCoordinator struct {
	mu             sync.RWMutex
	partyID        string
	leaderID       string
	members        map[string]PartyMember
	soloKillRate   float64 // Teto seguro de kills/min de um jogador solo
	bonusExpRate   float64 // Bônus de EXP cooperativo (ex: 0.45 para +45%)
	bestiaryKills  map[string]map[string]int
	lockedHunts    map[string]time.Time
}

// NewPartyCoordinator inicializa o coordenador
func NewPartyCoordinator(partyID, leaderID string, soloRate, bonusExp float64) *PartyCoordinator {
	if soloRate <= 0 {
		soloRate = 40.0 // Padrão seguro
	}
	if bonusExp <= 0 {
		bonusExp = 0.45 // +45% de EXP de grupo
	}

	return &PartyCoordinator{
		partyID:       partyID,
		leaderID:      leaderID,
		members:       make(map[string]PartyMember),
		soloKillRate:  soloRate,
		bonusExpRate:  bonusExp,
		bestiaryKills: make(map[string]map[string]int),
		lockedHunts:   make(map[string]time.Time),
	}
}

// RegisterMember adiciona ou atualiza um bot no grupo
func (pc *PartyCoordinator) RegisterMember(m PartyMember) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	m.LastSeen = time.Now()
	pc.members[m.ID] = m
}

// RemoveMember remove um bot que desconectou
func (pc *PartyCoordinator) RemoveMember(memberID string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	delete(pc.members, memberID)
	delete(pc.bestiaryKills, memberID)
}

// IsLeader verifica se a conta é o líder do grupo
func (pc *PartyCoordinator) IsLeader(accountID string) bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.leaderID == accountID
}

// MinMemberLevel retorna o menor nível entre os membros da party (para seleção de mapa seguro para todos)
func (pc *PartyCoordinator) MinMemberLevel() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	if len(pc.members) == 0 {
		return 1
	}
	minLv := math.MaxInt32
	for _, m := range pc.members {
		if m.Level > 0 && m.Level < minLv {
			minLv = m.Level
		}
	}
	if minLv == math.MaxInt32 {
		return 1
	}
	return minLv
}

// CanFormPartyWithLevelGap verifica se a diferença entre o maior e o menor nível respeita o limite do jogo (ex: <= 25 no Leveling Idle)
func (pc *PartyCoordinator) CanFormPartyWithLevelGap(maxGap int) bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	if len(pc.members) <= 1 || maxGap <= 0 {
		return true
	}
	minLv, maxLv := math.MaxInt32, 0
	for _, m := range pc.members {
		if m.Level < minLv {
			minLv = m.Level
		}
		if m.Level > maxLv {
			maxLv = m.Level
		}
	}
	return (maxLv - minLv) <= maxGap
}

// RecordMemberBestiary atualiza o progresso de abates de bestiário de um membro específico
func (pc *PartyCoordinator) RecordMemberBestiary(memberID, mobID string, kills int) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	if pc.bestiaryKills[memberID] == nil {
		pc.bestiaryKills[memberID] = make(map[string]int)
	}
	pc.bestiaryKills[memberID][mobID] = kills
}

// GetSharedBestiaryKills retorna o MENOR número de abates do monstro entre todos os membros da party
// (garante que o grupo só saia da rota quando todos os membros tiverem fechado o bestiário)
func (pc *PartyCoordinator) GetSharedBestiaryKills(mobID string) int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	if len(pc.members) == 0 {
		return 0
	}
	minKills := math.MaxInt32
	for memberID := range pc.members {
		k := 0
		if mMap, ok := pc.bestiaryKills[memberID]; ok {
			k = mMap[mobID]
		}
		if k < minKills {
			minKills = k
		}
	}
	if minKills == math.MaxInt32 {
		return 0
	}
	return minKills
}

// LockHunt adiciona uma área na blacklist temporária (ex: 15 min quando o servidor recusa por gate.locked / level_too_low)
func (pc *PartyCoordinator) LockHunt(areaID string, duration time.Duration) {
	if duration <= 0 {
		duration = 15 * time.Minute
	}
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.lockedHunts[areaID] = time.Now().Add(duration)
}

// IsHuntLocked verifica se a área está temporariamente bloqueada para a party
func (pc *PartyCoordinator) IsHuntLocked(areaID string) bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	until, ok := pc.lockedHunts[areaID]
	if !ok {
		return false
	}
	return time.Now().Before(until)
}

// CalculateThrottledKillRate calcula a taxa máxima segura de abates por minuto para cada membro individual
// Aplica a fórmula obrigatória da Regra 2 do AGENTS.md:
// KillRate_indiv <= SoloRate / (N * (1 + Bonus))
func (pc *PartyCoordinator) CalculateThrottledKillRate() float64 {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	size := len(pc.members)
	if size <= 1 {
		return pc.soloKillRate
	}

	bonusPercent := pc.bonusExpRate
	if bonusPercent <= 1.0 && bonusPercent > 0 {
		bonusPercent *= 100.0
	}
	return engine.CalculatePartyMaxKillRate(pc.soloKillRate, size, bonusPercent)
}

// GetSafetyIntervalSec devolve o intervalo mínimo em segundos entre abates para não estourar o anti-cheat
func (pc *PartyCoordinator) GetSafetyIntervalSec() float64 {
	kpm := pc.CalculateThrottledKillRate()
	if kpm <= 0 {
		return 1.0
	}
	return 60.0 / kpm
}

// String resumo para telemetria
func (pc *PartyCoordinator) String() string {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return fmt.Sprintf("Party '%s' (%d membros, MinLv %d) | Líder: %s | Max Kills/min por bot: %.2f (%.1fs/kill)",
		pc.partyID, len(pc.members), pc.MinMemberLevel(), pc.leaderID, pc.CalculateThrottledKillRate(), pc.GetSafetyIntervalSec())
}
