package subsystem

import (
	"sync"
)

// InventoryItem representa um item armazenado no inventário/mochila
type InventoryItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Count     int    `json:"count"`
	SellValue int    `json:"sellValue"`
	IsTrash   bool   `json:"isTrash"`
	Tier      int    `json:"tier,omitempty"`
}

// SellBatch representa o lote de itens seguro para venda em NPC
type SellBatch struct {
	Items     []InventoryItem
	TotalGold int64
}

// InventoryManager gerencia a capacidade da mochila e automatiza o auto-sell de lixo
type InventoryManager struct {
	mu             sync.RWMutex
	protectedItems map[string]bool
	itemPrices     map[string]int
	maxSlots       int
}

// NewInventoryManager inicializa o gerenciador com uma lista opcional de itens protegidos
func NewInventoryManager(maxSlots int, protected []string) *InventoryManager {
	im := &InventoryManager{
		protectedItems: make(map[string]bool),
		itemPrices:     make(map[string]int),
		maxSlots:       maxSlots,
	}

	for _, p := range protected {
		im.protectedItems[p] = true
	}
	return im
}

// Protect marca um item como protegido (nunca vender no auto-sell)
func (im *InventoryManager) Protect(itemID string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.protectedItems[itemID] = true
}

// SetPrice cadastra o valor de venda de um item no NPC
func (im *InventoryManager) SetPrice(itemID string, price int) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.itemPrices[itemID] = price
}

// ShouldSell avalia se a mochila está cheia ou atingiu o limiar de segurança (ex: 80% da capacidade)
func (im *InventoryManager) ShouldSell(usedSlots int, thresholdPercent float64) bool {
	if im.maxSlots <= 0 {
		return false
	}
	if thresholdPercent <= 0 {
		thresholdPercent = 0.80 // Padrão: 80% de capacidade ocupada
	}
	ratio := float64(usedSlots) / float64(im.maxSlots)
	return ratio >= thresholdPercent
}

// GetSellableBatch filtra a lista de itens e devolve somente os itens de lixo/venda segura
func (im *InventoryManager) GetSellableBatch(stash map[string]int) SellBatch {
	im.mu.RLock()
	defer im.mu.RUnlock()

	var batch SellBatch
	for id, count := range stash {
		if count <= 0 {
			continue
		}
		// Itens protegidos nunca são vendidos
		if im.protectedItems[id] {
			continue
		}
		// Deve ter preço de venda cadastrado > 0
		price, hasPrice := im.itemPrices[id]
		if !hasPrice || price <= 0 {
			continue
		}

		batch.Items = append(batch.Items, InventoryItem{
			ID:        id,
			Count:     count,
			SellValue: price,
			IsTrash:   true,
		})
		batch.TotalGold += int64(price * count)
	}

	return batch
}
