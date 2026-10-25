package subsystem

import (
	"sort"
	"sync"
)

// PotionDefinition define o poder de cura e custo de uma poção
type PotionDefinition struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	HealPower int    `json:"healPower"`
	CostGold  int    `json:"costGold"`
	IsMana    bool   `json:"isMana"`
}

// PotionManager otimiza o consumo e reposição de consumíveis sem desperdício de ouro
type PotionManager struct {
	mu           sync.RWMutex
	hpPotions    []PotionDefinition
	manaPotions  []PotionDefinition
	minStockKeep int
}

// NewPotionManager cria o gerenciador com as poções disponíveis no jogo
func NewPotionManager(potions []PotionDefinition, minStock int) *PotionManager {
	if minStock <= 0 {
		minStock = 20
	}

	var hp, mana []PotionDefinition
	for _, p := range potions {
		if p.IsMana {
			mana = append(mana, p)
		} else {
			hp = append(hp, p)
		}
	}

	// Ordenar por poder de cura decrescente
	sort.Slice(hp, func(i, j int) bool { return hp[i].HealPower > hp[j].HealPower })
	sort.Slice(mana, func(i, j int) bool { return mana[i].HealPower > mana[j].HealPower })

	return &PotionManager{
		hpPotions:    hp,
		manaPotions:  mana,
		minStockKeep: minStock,
	}
}

// SelectBestHPPotion escolhe a poção ideal para a quantidade de vida perdida
// Evita usar poções grandes quando o dano sofrido for pequeno
func (pm *PotionManager) SelectBestHPPotion(missingHP int, stash map[string]int) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, p := range pm.hpPotions {
		if stash[p.ID] > 0 && missingHP >= (p.HealPower/2) {
			return p.ID
		}
	}

	// Fallback para menor poção disponível em estoque
	for i := len(pm.hpPotions) - 1; i >= 0; i-- {
		p := pm.hpPotions[i]
		if stash[p.ID] > 0 {
			return p.ID
		}
	}
	return ""
}

// NeedsRestock avalia se o estoque de qualquer poção essencial caiu abaixo do limiar
func (pm *PotionManager) NeedsRestock(stash map[string]int) []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var toBuy []string
	for _, p := range pm.hpPotions {
		if stash[p.ID] < pm.minStockKeep {
			toBuy = append(toBuy, p.ID)
		}
	}
	return toBuy
}
