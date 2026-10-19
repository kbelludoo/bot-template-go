package engine

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// StaminaType define o modelo de métrica de tempo/esforço do jogo
type StaminaType int

const (
	// StaminaTypeDecreasing: Stamina decrescente (ex: Tibia, Huntera, Baiak, 42h -> 0h).
	StaminaTypeDecreasing StaminaType = iota
	// StaminaTypeIncreasingFatigue: Fadiga crescente (ex: Leveling Idle, Protobuf FatigueState, 0 -> 1000 milli = 0% -> 100%).
	StaminaTypeIncreasingFatigue
)

// StaminaConfig encapsula os limiares de segurança para prevenção de desperdício de EXP
type StaminaConfig struct {
	Type StaminaType

	// Para Fadiga Crescente (0.0% a 100.0%)
	MaxFatiguePercent    float64 // Limiar crítico de parada (Padrão: 80.0%)
	ResumeFatiguePercent float64 // Limiar seguro para retomar caçada (Padrão: 30.0%)

	// Para Farm Disponível (milissegundos)
	MinFarmAvailableMs    uint64 // Parar se farm restante < valor (Padrão: 120000 = 2 min)
	ResumeFarmAvailableMs uint64 // Retomar se farm restante > valor (Padrão: 600000 = 10 min)

	// Para Stamina Decrescente (horas)
	MinStaminaHours    float64 // Parar caçada ao atingir limite crítico (Padrão: 14.0h)
	ResumeStaminaHours float64 // Retomar caçada ao atingir stamina verde (Padrão: 38.0h)

	// Histerese temporal obrigatória (tempo mínimo de descanso contínuo)
	MinRestDuration time.Duration // Padrão: 5 minutos
}

// DefaultStaminaConfig retorna os parâmetros universais anti-desperdício
func DefaultStaminaConfig(t StaminaType) StaminaConfig {
	return StaminaConfig{
		Type:                  t,
		MaxFatiguePercent:    80.0,
		ResumeFatiguePercent: 30.0,
		MinFarmAvailableMs:    120000, // 2 min
		ResumeFarmAvailableMs: 600000, // 10 min
		MinStaminaHours:       14.0,   // limite onde a EXP cai para 50% ou 0%
		ResumeStaminaHours:    38.0,   // stamina verde com bônus
		MinRestDuration:       5 * time.Minute,
	}
}

// StaminaController administra a transição segura entre caça e descanso
type StaminaController struct {
	mu            sync.RWMutex
	cfg           StaminaConfig
	isResting     bool
	restStartTime time.Time

	fatigueMilli    uint32
	farmAvailableMs uint64
	staminaHours    float64
	onlineKpm       float64
}

// NewStaminaController inicializa um controlador thread-safe
func NewStaminaController(cfg StaminaConfig) *StaminaController {
	return &StaminaController{
		cfg: cfg,
	}
}

// UpdateFatigue atualiza métricas de fadiga (comum em protocolos binários/protobuf)
func (s *StaminaController) UpdateFatigue(milli uint32, farmAvailMs uint64, kpm float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fatigueMilli = milli
	s.farmAvailableMs = farmAvailMs
	s.onlineKpm = kpm
}

// UpdateStaminaHours atualiza métrica de horas de stamina (estilo Tibia/Huntera)
func (s *StaminaController) UpdateStaminaHours(hours float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.staminaHours = hours
}

// ShouldRest avalia se a conta deve parar imediatamente de caçar
func (s *StaminaController) ShouldRest() (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.isResting {
		return false, ""
	}

	switch s.cfg.Type {
	case StaminaTypeIncreasingFatigue:
		fatiguePct := float64(s.fatigueMilli) / 10.0
		if fatiguePct >= s.cfg.MaxFatiguePercent {
			return true, fmt.Sprintf("Fadiga crítica atingida: %.1f%% (limite: %.1f%%)", fatiguePct, s.cfg.MaxFatiguePercent)
		}
		if s.farmAvailableMs > 0 && s.farmAvailableMs < s.cfg.MinFarmAvailableMs {
			availMin := float64(s.farmAvailableMs) / 60000.0
			return true, fmt.Sprintf("Tempo de farm esgotando: %.1f min restantes", availMin)
		}

	case StaminaTypeDecreasing:
		if s.staminaHours > 0 && s.staminaHours <= s.cfg.MinStaminaHours {
			return true, fmt.Sprintf("Stamina crítica atingida: %.1fh (limite: %.1fh)", s.staminaHours, s.cfg.MinStaminaHours)
		}
	}

	return false, ""
}

// CanResume avalia se a conta já descansou o suficiente e pode voltar à caçada
func (s *StaminaController) CanResume() (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.isResting {
		return false, ""
	}

	restDur := time.Since(s.restStartTime)
	if restDur < s.cfg.MinRestDuration {
		remSec := (s.cfg.MinRestDuration - restDur).Seconds()
		return false, fmt.Sprintf("Histerese de descanso ativa (aguarde mais %.0fs)", remSec)
	}

	switch s.cfg.Type {
	case StaminaTypeIncreasingFatigue:
		fatiguePct := float64(s.fatigueMilli) / 10.0
		farmRecovered := s.farmAvailableMs == 0 || s.farmAvailableMs >= s.cfg.ResumeFarmAvailableMs
		fatigueRecovered := fatiguePct <= s.cfg.ResumeFatiguePercent

		if fatigueRecovered && farmRecovered {
			return true, fmt.Sprintf("Fadiga recuperada: %.1f%% | Farm disponível: %.1f min", fatiguePct, float64(s.farmAvailableMs)/60000.0)
		}
		return false, fmt.Sprintf("Ainda recuperando (Fadiga: %.1f%% / Farm: %.1f min)", fatiguePct, float64(s.farmAvailableMs)/60000.0)

	case StaminaTypeDecreasing:
		if s.staminaHours >= s.cfg.ResumeStaminaHours {
			return true, fmt.Sprintf("Stamina verde recuperada: %.1fh", s.staminaHours)
		}
		return false, fmt.Sprintf("Ainda recuperando stamina (%.1fh / meta: %.1fh)", s.staminaHours, s.cfg.ResumeStaminaHours)
	}

	return true, "Descanso concluído"
}

// EnterRest sinaliza a transição para modo de descanso / zona segura
func (s *StaminaController) EnterRest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isResting = true
	s.restStartTime = time.Now()
}

// ExitRest encerra o modo descanso
func (s *StaminaController) ExitRest() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isResting = false
}

// IsResting retorna se o bot está atualmente descansando
func (s *StaminaController) IsResting() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isResting
}

// ExportExtra gera o dicionário com campos padronizados para status.Extra
func (s *StaminaController) ExportExtra() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	extra := make(map[string]any)
	extra["isResting"] = s.isResting

	if s.isResting {
		restMin := time.Since(s.restStartTime).Minutes()
		extra["restingMinutes"] = math.Round(restMin*10) / 10
	}

	switch s.cfg.Type {
	case StaminaTypeIncreasingFatigue:
		extra["fatiguePercent"] = math.Round((float64(s.fatigueMilli)/10.0)*10) / 10
		extra["farmAvailableMin"] = math.Round((float64(s.farmAvailableMs)/60000.0)*10) / 10
		extra["onlineKpm"] = math.Round(s.onlineKpm*10) / 10

	case StaminaTypeDecreasing:
		extra["staminaHours"] = math.Round(s.staminaHours*10) / 10
	}

	return extra
}
