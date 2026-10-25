package subsystem

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// AreaCandidate representa uma área ou rota de caça avaliada pelo solver
type AreaCandidate struct {
	ID           string   // Slug ou ID da rota (ex: "undead_dragon", "mata_larvas")
	Name         string   // Nome de exibição
	MinLevel     int      // Nível mínimo recomendado
	MaxLevel     int      // Nível máximo onde a EXP é plena
	BaseExp      float64  // EXP base concedida pelos monstros
	BaseGold     float64  // Ouro médio por abate
	KillsPerHour float64  // Taxa estimada de abates/hora
	Monsters     []string // Espécies presentes
	Tier         int      // Dificuldade da área
}

// AreaBenchmark armazena métricas empíricas reais medidas durante a caça (padrão Baiak Idle)
type AreaBenchmark struct {
	TotalExp      float64
	TotalGold     float64
	TotalKills    int
	TotalDuration time.Duration
}

// ExpPerHour retorna a EXP/h empírica medida na área
func (b AreaBenchmark) ExpPerHour() float64 {
	hours := b.TotalDuration.Hours()
	if hours <= 0 {
		return 0
	}
	return b.TotalExp / hours
}

// SolverCriteria parâmetros de entrada do personagem para o solver tomar a melhor decisão
type SolverCriteria struct {
	CharacterLevel          int
	PrimaryAttackType       string
	ActiveWeaponTier        int
	TargetBestiary          map[string]bool                       // Monstros que ainda precisam de abates para o bestiário
	BestiaryRemainingKills  map[string]int                        // Abates restantes por monstro para cálculo de urgência quadrática
	ElementalMatrix         func(attackType, monster string) float64
	DefensiveRiskMultiplier func(area AreaCandidate) float64      // Multiplicador de dano sofrido (bloqueia se >= 2.0 em nível alto)
}

// SolvedRoute resultado calculado pelo solver
type SolvedRoute struct {
	Area       AreaCandidate
	Score      float64
	Reason     string
	IsBestiary bool
}

// RouteSolver calcula matematicamente a rota mais eficiente de caça combinando teoria e telemetria empírica
type RouteSolver struct {
	mu         sync.RWMutex
	areas      []AreaCandidate
	benchmarks map[string]AreaBenchmark
}

// NewRouteSolver inicializa o solucionador com a lista de áreas disponíveis
func NewRouteSolver(areas []AreaCandidate) *RouteSolver {
	return &RouteSolver{
		areas:      areas,
		benchmarks: make(map[string]AreaBenchmark),
	}
}

// RecordBenchmark registra dados reais de EXP/ouro/abates medidos em uma sessão de caça
func (s *RouteSolver) RecordBenchmark(areaID string, expGained, goldGained float64, kills int, duration time.Duration) {
	if duration <= 0 || areaID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	b := s.benchmarks[areaID]
	b.TotalExp += expGained
	b.TotalGold += goldGained
	b.TotalKills += kills
	b.TotalDuration += duration
	s.benchmarks[areaID] = b
}

// FindOptimalRoute avalia todas as áreas candidatas e retorna a mais lucrativa e segura
func (s *RouteSolver) FindOptimalRoute(crit SolverCriteria) (SolvedRoute, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.areas) == 0 {
		return SolvedRoute{}, false
	}

	var best SolvedRoute
	best.Score = -1.0

	for _, area := range s.areas {
		// Regra de segurança: personagem não pode caçar em área acima do seu nível
		if area.MinLevel > crit.CharacterLevel {
			continue
		}

		// Anti-Suicídio Elemental: se o risco defensivo for >= 2.0x em área de nível próximo/alto, bloquear
		defRisk := 1.0
		if crit.DefensiveRiskMultiplier != nil {
			defRisk = crit.DefensiveRiskMultiplier(area)
			if defRisk >= 2.0 && crit.CharacterLevel < area.MaxLevel+15 {
				continue
			}
		}

		// Penalidade de nível caso a área seja muito fraca
		levelDiff := crit.CharacterLevel - area.MaxLevel
		levelPenalty := 1.0
		if levelDiff > 30 {
			levelPenalty = 0.4
		} else if levelDiff > 15 {
			levelPenalty = 0.7
		}

		// Cálculo de vantagem elemental média ofensiva
		avgElementalMultiplier := 1.0
		if crit.ElementalMatrix != nil && len(area.Monsters) > 0 {
			var sum float64
			for _, m := range area.Monsters {
				sum += crit.ElementalMatrix(crit.PrimaryAttackType, m)
			}
			avgElementalMultiplier = sum / float64(len(area.Monsters))
		}

		// Bônus de Bestiário e Urgência Quadrática (padrão PokeIdle)
		hasPendingBestiary := false
		bestiaryBonus := 1.0
		if crit.TargetBestiary != nil {
			for _, m := range area.Monsters {
				if crit.TargetBestiary[m] {
					hasPendingBestiary = true
					bestiaryBonus = 1.75 // +75% de prioridade base para fechar códex
					if crit.BestiaryRemainingKills != nil {
						if rem, ok := crit.BestiaryRemainingKills[m]; ok && rem > 0 {
							urgency := (1000.0 / float64(rem+10)) * math.Pow(avgElementalMultiplier, 2)
							if urgency > bestiaryBonus {
								bestiaryBonus = urgency
							}
						}
					}
					break
				}
			}
		}

		// Produtividade: se houver benchmark empírico consolidado (>= 2 min), usar EXP/h real medida
		baseThroughput := area.KillsPerHour * math.Max(1.0, area.BaseExp) * avgElementalMultiplier
		if bm, ok := s.benchmarks[area.ID]; ok && bm.TotalDuration >= 2*time.Minute && bm.ExpPerHour() > 0 {
			baseThroughput = bm.ExpPerHour()
		}

		avgLevel := float64(area.MinLevel+area.MaxLevel) / 2.0
		safetyFactor := 1.0 / math.Max(0.5, defRisk)
		score := baseThroughput * levelPenalty * (1.0 + (avgLevel * 0.02)) * bestiaryBonus * safetyFactor

		if score > best.Score {
			best.Score = score
			best.Area = area
			best.IsBestiary = hasPendingBestiary

			bestiaryTag := ""
			if hasPendingBestiary {
				bestiaryTag = " | 📖 Bestiário Pendente"
			}
			best.Reason = fmt.Sprintf("Lv %d-%d | Eficácia: %.1fx | Score: %.0f%s",
				area.MinLevel, area.MaxLevel, avgElementalMultiplier, score, bestiaryTag)
		}
	}

	if best.Score < 0 {
		return SolvedRoute{
			Area:   s.areas[0],
			Score:  0,
			Reason: "Fallback para área inicial",
		}, true
	}

	return best, true
}

// CombatSimInput parâmetros de simulação de sobrevivência e TTK (padrão Sigma Idle)
type CombatSimInput struct {
	PlayerHP             float64
	PlayerATK            float64
	PlayerDEF            float64
	PlayerAtkIntervalSec float64
	PlayerHealPerSec     float64 // Cura combinada de regen + poções/s
	MobHP                float64
	MobATK               float64
	MobDEF               float64
	MobAtkIntervalSec    float64
	MobDensity           int // Quantidade média de monstros batendo simultaneamente
}

// CombatSimResult diagnóstico matemático de combate
type CombatSimResult struct {
	PlayerDPS   float64
	IncomingDPS float64
	TTKSec      float64
	CanTank     bool
	KillsPerMin float64
}

// SimulateCombat calcula TTK, DPS líquido recebido e se o personagem aguenta tankar a área
func SimulateCombat(in CombatSimInput) CombatSimResult {
	atkInterval := math.Max(0.2, in.PlayerAtkIntervalSec)
	mobInterval := math.Max(0.5, in.MobAtkIntervalSec)
	density := math.Max(1.0, float64(in.MobDensity))

	dmgPerHit := math.Max(1.0, in.PlayerATK-in.MobDEF)
	playerDPS := dmgPerHit / atkInterval
	ttk := math.Max(0.1, in.MobHP/playerDPS)

	mobHit := math.Max(1.0, in.MobATK-in.PlayerDEF)
	incomingDPS := (mobHit / mobInterval) * density

	netDrainPerSec := incomingDPS - in.PlayerHealPerSec
	canTank := true
	if netDrainPerSec > 0 {
		// Precisa sobreviver a pelo menos 10 abates consecutivos sem zerar o HP
		timeToDie := in.PlayerHP / netDrainPerSec
		if timeToDie < ttk*10.0 {
			canTank = false
		}
	}

	// Adiciona 1.0s de transição/spawn entre monstros
	kpm := 60.0 / (ttk + 1.0)
	return CombatSimResult{
		PlayerDPS:   playerDPS,
		IncomingDPS: incomingDPS,
		TTKSec:      ttk,
		CanTank:     canTank,
		KillsPerMin: kpm,
	}
}

var pokemonTypeChart = map[string]map[string]float64{
	"normal":   {"rock": 0.5, "ghost": 0.0, "steel": 0.5},
	"fire":     {"fire": 0.5, "water": 0.5, "grass": 2.0, "ice": 2.0, "bug": 2.0, "rock": 0.5, "dragon": 0.5, "steel": 2.0},
	"water":    {"fire": 2.0, "water": 0.5, "grass": 0.5, "ground": 2.0, "rock": 2.0, "dragon": 0.5},
	"electric": {"water": 2.0, "electric": 0.5, "grass": 0.5, "ground": 0.0, "flying": 2.0, "dragon": 0.5},
	"grass":    {"fire": 0.5, "water": 2.0, "grass": 0.5, "poison": 0.5, "ground": 2.0, "flying": 0.5, "bug": 0.5, "rock": 2.0, "dragon": 0.5, "steel": 0.5},
	"ice":      {"fire": 0.5, "water": 0.5, "grass": 2.0, "ice": 0.5, "ground": 2.0, "flying": 2.0, "dragon": 2.0, "steel": 0.5},
	"fighting": {"normal": 2.0, "ice": 2.0, "poison": 0.5, "flying": 0.5, "psychic": 0.5, "bug": 0.5, "rock": 2.0, "ghost": 0.0, "dark": 2.0, "steel": 2.0, "fairy": 0.5},
	"poison":   {"grass": 2.0, "poison": 0.5, "ground": 0.5, "rock": 0.5, "ghost": 0.5, "steel": 0.0, "fairy": 2.0},
	"ground":   {"fire": 2.0, "electric": 2.0, "grass": 0.5, "poison": 2.0, "flying": 0.0, "bug": 0.5, "rock": 2.0, "steel": 2.0},
	"flying":   {"electric": 0.5, "grass": 2.0, "fighting": 2.0, "bug": 2.0, "rock": 0.5, "steel": 0.5},
	"psychic":  {"fighting": 2.0, "poison": 2.0, "psychic": 0.5, "dark": 0.0, "steel": 0.5},
	"bug":      {"fire": 0.5, "grass": 2.0, "fighting": 0.5, "poison": 0.5, "flying": 0.5, "psychic": 2.0, "ghost": 0.5, "dark": 2.0, "steel": 0.5, "fairy": 0.5},
	"rock":     {"fire": 2.0, "ice": 2.0, "fighting": 0.5, "ground": 0.5, "flying": 2.0, "bug": 2.0, "steel": 0.5},
	"ghost":    {"normal": 0.0, "psychic": 2.0, "ghost": 2.0, "dark": 0.5},
	"dragon":   {"dragon": 2.0, "steel": 0.5, "fairy": 0.0},
	"dark":     {"fighting": 0.5, "psychic": 2.0, "ghost": 2.0, "dark": 0.5, "fairy": 0.5},
	"steel":    {"fire": 0.5, "water": 0.5, "electric": 0.5, "ice": 2.0, "rock": 2.0, "steel": 0.5, "fairy": 2.0},
	"fairy":    {"fire": 0.5, "fighting": 2.0, "poison": 0.5, "dragon": 2.0, "dark": 2.0, "steel": 0.5},
}

// PokemonElementalEffectiveness retorna o multiplicador oficial da matriz 18x18 de tipos Pokémon (0.0, 0.5, 1.0, 2.0)
func PokemonElementalEffectiveness(attackType, defenseType string) float64 {
	atk := strings.ToLower(strings.TrimSpace(attackType))
	def := strings.ToLower(strings.TrimSpace(defenseType))
	if row, ok := pokemonTypeChart[atk]; ok {
		if mult, exists := row[def]; exists {
			return mult
		}
	}
	return 1.0
}
