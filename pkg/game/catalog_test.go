package game

import (
	"testing"
)

func TestCatalogAndEffectiveness(t *testing.T) {
	InitCatalog()

	// 1. Testar tabela de tipos e fraquezas
	// Água contra Fogo deve ser 2.0x
	if eff := GetEffectiveness("water", []string{"fire"}); eff != 2.0 {
		t.Errorf("Esperado efetividade 2.0 para water -> fire, obteve %f", eff)
	}

	// Elétrico contra Água e Voador (dupla fraqueza, ex: Gyarados) deve ser 4.0x
	if eff := GetEffectiveness("electric", []string{"water", "flying"}); eff != 4.0 {
		t.Errorf("Esperado efetividade 4.0 para electric -> water/flying, obteve %f", eff)
	}

	// Fantasma contra Normal deve ser 0x (imunidade)
	if eff := GetEffectiveness("ghost", []string{"normal"}); eff != 0.0 {
		t.Errorf("Esperado efetividade 0.0 para ghost -> normal, obteve %f", eff)
	}

	// Planta contra Fogo deve ser 0.5x
	if eff := GetEffectiveness("grass", []string{"fire"}); eff != 0.5 {
		t.Errorf("Esperado efetividade 0.5 para grass -> fire, obteve %f", eff)
	}
}

func TestOptimalHuntSolver(t *testing.T) {
	InitCatalog()

	// Simular um time com um Charmander (tipo fogo) nível 15
	slotZero := 0
	team := []PokemonState{
		{
			ID:       1,
			Especie:  "Charmander",
			Nivel:    15,
			TeamSlot: &slotZero,
		},
	}

	bestHunt, reason := FindBestHunt(team, 15)
	if bestHunt.Slug == "" {
		t.Fatalf("Nenhuma hunt encontrada pelo solver!")
	}
	t.Logf("Melhor rota encontrada: %s (%s) - %s", bestHunt.Nome, bestHunt.Slug, reason)

	// Contra insetos e plantas (Mata das Larvas ou Clareira dos Brotos), Charmander tem super efetivo
	if len(bestHunt.Pokes) == 0 {
		t.Errorf("Hunt retornada não tem monstros definidos")
	}
}

func TestAutoSellLootFilter(t *testing.T) {
	InitCatalog()

	stash := map[string]int{
		"seed":           50,  // Loot normal (deve ser vendido)
		"pair_of_leaves": 20,  // Loot normal (deve ser vendido)
		"fire_stone":     3,   // Pedra de evolução (NUNCA vender)
		"candy":          5,   // Rare candy (NUNCA vender)
		"pokeball":       100, // Pokébola (NUNCA vender)
		"potion":         10,  // Poção (NUNCA vender)
		"devoted_token":  40,  // Token (NUNCA vender)
	}

	items := GetSellableLoot(stash)
	if len(items) == 0 {
		t.Fatalf("Esperava itens de loot para vender, retornou vazio")
	}

	for _, it := range items {
		if ProtectedItems[it.Item] {
			t.Errorf("Item protegido '%s' foi marcado para venda!", it.Item)
		}
	}
	t.Logf("Itens seguros para venda: %+v", items)
}

func TestPotionSelector(t *testing.T) {
	stash := map[string]int{
		"potion":       5,
		"great_potion": 3,
		"ultra_potion": 1,
	}

	// Se perdeu 180 de HP, great_potion (150) é a mais indicada
	p := GetBestPotion(180, stash)
	if p != "great_potion" && p != "ultra_potion" {
		t.Errorf("Esperava great_potion ou ultra_potion, obteve: %s", p)
	}

	// Se perdeu 400 de HP, ultra_potion (400) é a mais indicada
	p2 := GetBestPotion(400, stash)
	if p2 != "ultra_potion" {
		t.Errorf("Esperava ultra_potion, obteve: %s", p2)
	}
}
