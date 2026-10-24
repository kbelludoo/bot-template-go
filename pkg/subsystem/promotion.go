package subsystem

import (
	"sync"
)

// PromotionTier representa um degrau de promoção de classe
type PromotionTier struct {
	LevelReq   int    `json:"levelReq"`
	CostGold   int64  `json:"costGold"`
	TargetName string `json:"targetName"`
	ActionCode string `json:"actionCode"`
}

// BlessingConfig define os requisitos de bênçãos de proteção contra morte
type BlessingConfig struct {
	MaxBlessings int   `json:"maxBlessings"` // Geralmente 5
	CostPerBless int64 `json:"costPerBless"`
}

// PromotionManager gerencia promoções de vocação e seguro contra mortes (blessings)
type PromotionManager struct {
	mu           sync.RWMutex
	tiers        []PromotionTier
	blessCfg     BlessingConfig
	isPromoted   bool
	currentBless int
}

// NewPromotionManager inicializa o módulo de promoção
func NewPromotionManager(tiers []PromotionTier, bCfg BlessingConfig) *PromotionManager {
	return &PromotionManager{
		tiers:    tiers,
		blessCfg: bCfg,
	}
}

// CheckEligiblePromotion verifica se o personagem atingiu nível e ouro para comprar uma nova promoção
func (pm *PromotionManager) CheckEligiblePromotion(currentLevel int, currentGold int64, currentTitle string) (PromotionTier, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, tier := range pm.tiers {
		if currentLevel >= tier.LevelReq && currentGold >= tier.CostGold && currentTitle != tier.TargetName {
			return tier, true
		}
	}
	return PromotionTier{}, false
}

// MissingBlessings calcula quantas bênçãos faltam para 100% de proteção contra perda de EXP/itens
func (pm *PromotionManager) MissingBlessings(activeBlessCount int) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if activeBlessCount < pm.blessCfg.MaxBlessings {
		return pm.blessCfg.MaxBlessings - activeBlessCount
	}
	return 0
}

// CanAffordBlessings calcula se o ouro acumulado paga as bênçãos pendentes
func (pm *PromotionManager) CanAffordBlessings(activeBlessCount int, gold int64) bool {
	missing := pm.MissingBlessings(activeBlessCount)
	if missing <= 0 {
		return false
	}
	totalCost := int64(missing) * pm.blessCfg.CostPerBless
	return gold >= totalCost
}
