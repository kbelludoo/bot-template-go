package engine

import (
	"hash/fnv"
	"math"
	"sync"
	"time"
)

// AdaptiveSpeedConfig configura o auto-ajuste de velocidade de ciclo/run (padrão Farm Ricochet)
type AdaptiveSpeedConfig struct {
	InitialSpeed      float64
	MinSpeed          float64
	MaxSpeed          float64
	StepUp            float64
	StepDown          float64
	CleanRunsRequired int
}

// DefaultAdaptiveSpeedConfig retorna parâmetros conservadores de aceleração adaptativa
func DefaultAdaptiveSpeedConfig() AdaptiveSpeedConfig {
	return AdaptiveSpeedConfig{
		InitialSpeed:      1.20,
		MinSpeed:          1.00,
		MaxSpeed:          2.50,
		StepUp:            0.05,
		StepDown:          0.10,
		CleanRunsRequired: 3,
	}
}

// AdaptiveSpeedTuner ajusta dinamicamente o multiplicador de velocidade conforme aceitação do servidor
type AdaptiveSpeedTuner struct {
	mu        sync.RWMutex
	cfg       AdaptiveSpeedConfig
	current   float64
	cleanRuns int
}

// NewAdaptiveSpeedTuner cria um ajustador de velocidade adaptativo
func NewAdaptiveSpeedTuner(cfg AdaptiveSpeedConfig) *AdaptiveSpeedTuner {
	if cfg.MinSpeed <= 0 {
		cfg.MinSpeed = 1.0
	}
	if cfg.MaxSpeed < cfg.MinSpeed {
		cfg.MaxSpeed = cfg.MinSpeed
	}
	if cfg.InitialSpeed < cfg.MinSpeed || cfg.InitialSpeed > cfg.MaxSpeed {
		cfg.InitialSpeed = cfg.MinSpeed
	}
	if cfg.StepUp <= 0 {
		cfg.StepUp = 0.05
	}
	if cfg.StepDown <= 0 {
		cfg.StepDown = 0.10
	}
	if cfg.CleanRunsRequired <= 0 {
		cfg.CleanRunsRequired = 3
	}
	return &AdaptiveSpeedTuner{
		cfg:     cfg,
		current: cfg.InitialSpeed,
	}
}

// RecordCleanRun registra uma run/ciclo aceito de primeira pelo servidor sem RetryAfterMs
func (t *AdaptiveSpeedTuner) RecordCleanRun() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cleanRuns++
	if t.cleanRuns >= t.cfg.CleanRunsRequired {
		t.cleanRuns = 0
		t.current = math.Min(t.cfg.MaxSpeed, math.Round((t.current+t.cfg.StepUp)*100)/100)
	}
	return t.current
}

// RecordThrottle registra que o servidor pediu RetryAfterMs ou rejeitou por velocidade
func (t *AdaptiveSpeedTuner) RecordThrottle() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cleanRuns = 0
	t.current = math.Max(t.cfg.MinSpeed, math.Round((t.current-t.cfg.StepDown)*100)/100)
	return t.current
}

// Speed retorna o multiplicador de velocidade atual
func (t *AdaptiveSpeedTuner) Speed() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current
}

// CalculateRunDuration calcula a duração segura de uma run baseada na taxa canônica do mapa
func (t *AdaptiveSpeedTuner) CalculateRunDuration(targetGoal, onlineRatePerSec, minDurationSec float64) time.Duration {
	if minDurationSec <= 0 {
		minDurationSec = 8.0
	}
	if onlineRatePerSec <= 0 || targetGoal <= 0 {
		return time.Duration(minDurationSec * float64(time.Second))
	}
	speed := t.Speed()
	sec := math.Max(minDurationSec, targetGoal/(onlineRatePerSec*speed))
	return time.Duration(sec * float64(time.Second))
}

// AntiWipeTracker monitora mortes em janela deslizante para rebaixar automaticamente a área de caça (padrão Baiak Idle)
type AntiWipeTracker struct {
	mu              sync.Mutex
	window          time.Duration
	restoreCooldown time.Duration
	maxDeaths       int
	deathTimes      []time.Time
	downgraded      bool
	downgradedAt    time.Time
}

// NewAntiWipeTracker inicializa o protetor contra sequência de mortes (default: 3 mortes em 10m -> downgrade por 10m)
func NewAntiWipeTracker(maxDeaths int, window, restoreCooldown time.Duration) *AntiWipeTracker {
	if maxDeaths <= 0 {
		maxDeaths = 3
	}
	if window <= 0 {
		window = 10 * time.Minute
	}
	if restoreCooldown <= 0 {
		restoreCooldown = 10 * time.Minute
	}
	return &AntiWipeTracker{
		window:          window,
		restoreCooldown: restoreCooldown,
		maxDeaths:       maxDeaths,
	}
}

// RecordDeath registra uma morte e informa se o bot deve rebaixar imediatamente para uma área mais fraca
func (a *AntiWipeTracker) RecordDeath(now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if now.IsZero() {
		now = time.Now()
	}

	cutoff := now.Add(-a.window)
	valid := a.deathTimes[:0]
	for _, ts := range a.deathTimes {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	valid = append(valid, now)
	a.deathTimes = valid

	if len(a.deathTimes) >= a.maxDeaths {
		a.downgraded = true
		a.downgradedAt = now
		return true
	}
	return a.downgraded
}

// IsDowngraded verifica se o modo de segurança pós-morte está ativo (restaurando se o cooldown já expirou sem novas mortes)
func (a *AntiWipeTracker) IsDowngraded(now time.Time) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.downgraded {
		return false
	}
	if now.IsZero() {
		now = time.Now()
	}
	if now.Sub(a.downgradedAt) >= a.restoreCooldown {
		a.downgraded = false
		a.deathTimes = nil
		return false
	}
	return true
}

// RankingStealthController controla o teto de nível relativo ao Top 1 do ranking público (padrão LoL Idle)
type RankingStealthController struct {
	mu                sync.RWMutex
	topRankingLevel   int
	safetyMargin      int
	stealthStartLevel int
	stealthKillFactor float64
}

// NewRankingStealthController cria o controlador furtivo de ranking
func NewRankingStealthController(safetyMargin, stealthStartLevel int, stealthKillFactor float64) *RankingStealthController {
	if safetyMargin <= 0 {
		safetyMargin = 5
	}
	if stealthStartLevel <= 0 {
		stealthStartLevel = 30
	}
	if stealthKillFactor <= 0 || stealthKillFactor >= 1.0 {
		stealthKillFactor = 0.5
	}
	return &RankingStealthController{
		safetyMargin:      safetyMargin,
		stealthStartLevel: stealthStartLevel,
		stealthKillFactor: stealthKillFactor,
	}
}

// UpdateTopRankingLevel atualiza o nível do líder do ranking público
func (r *RankingStealthController) UpdateTopRankingLevel(topLevel int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if topLevel > 0 {
		r.topRankingLevel = topLevel
	}
}

// ShouldFreezeXP informa se o bot deve congelar o ganho de XP para nunca aparecer no topo do ranking público
func (r *RankingStealthController) ShouldFreezeXP(currentLevel int) (bool, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.topRankingLevel <= r.safetyMargin {
		return false, 0
	}
	capLevel := r.topRankingLevel - r.safetyMargin
	return currentLevel >= capLevel, capLevel
}

// KillRateMultiplier retorna o fator de desaceleração furtiva (ex: 0.5x acima do nível 30)
func (r *RankingStealthController) KillRateMultiplier(currentLevel int) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if currentLevel >= r.stealthStartLevel {
		return r.stealthKillFactor
	}
	return 1.0
}

// CircularPatrolPosition calcula coordenadas (x, y) em órbita circular determinística por nick (padrão LoL Idle)
// Mantém todos os bots do grupo dentro do raio de bônus de party sem empilhar no mesmo pixel
func CircularPatrolPosition(centerX, centerY, radius, angularVelRadPerSec float64, nick string, now time.Time) (float64, float64) {
	if radius <= 0 {
		radius = 3.0
	}
	if angularVelRadPerSec == 0 {
		angularVelRadPerSec = 0.6
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(nick))
	phaseOffset := float64(h.Sum32()%6283) / 1000.0 // 0 .. 2*pi

	tSec := float64(now.UnixMilli()) / 1000.0
	angle := tSec*angularVelRadPerSec + phaseOffset

	x := centerX + math.Cos(angle)*radius
	y := centerY + math.Sin(angle)*radius
	return math.Round(x*100) / 100, math.Round(y*100) / 100
}
