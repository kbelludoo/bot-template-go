package game

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
)

//go:embed data/hunts.json
var huntsJSON []byte

//go:embed data/pokedex.json
var pokedexJSON []byte

//go:embed data/loot.json
var lootJSON []byte

// Hunt representa uma área de caça do jogo
type Hunt struct {
	Slug         string   `json:"slug"`
	Nome         string   `json:"nome"`
	Descricao    string   `json:"descricao"`
	Tier         int      `json:"tier"`
	Pokes        []string `json:"pokes"`
	Nivel        []int    `json:"nivel"` // [min, max]
	KillsPorHora int      `json:"killsPorHora"`
}

// PokedexData raiz de pokedex.json
type PokedexData struct {
	Tipos   map[string]map[string]float64 `json:"tipos"`
	Pokemon map[string]PokedexEntry       `json:"pokemon"`
}

type PokedexEntry struct {
	Num         int            `json:"num"`
	Tipos       []string       `json:"tipos"`
	Base        BaseStats      `json:"base"`
	Moves       []PokedexMove  `json:"moves"`
	CaptureRate int            `json:"capture_rate"`
	Exp         int            `json:"exp"`
	NivelReq    int            `json:"nivel_req"`
}

type BaseStats struct {
	HP  int `json:"hp"`
	Atk int `json:"atk"`
	Def int `json:"def"`
	SpA int `json:"spa"`
	SpD int `json:"spd"`
	Spe int `json:"spe"`
}

type PokedexMove struct {
	Nome    string `json:"nome"`
	Nivel   int    `json:"nivel"`
	Power   int    `json:"power"`
	Tipo    string `json:"tipo"`
	CD      int    `json:"cd"`
	Passiva bool   `json:"passiva"`
}

// LootData raiz de loot.json
type LootData struct {
	Itens map[string]LootItem `json:"itens"`
}

type LootItem struct {
	Nome  string `json:"nome"`
	Venda int    `json:"venda"`
	ID    int    `json:"id"`
}

type SellItem struct {
	Item string `json:"item"`
	Qtd  int    `json:"qtd"`
}

var (
	catalogOnce sync.Once
	huntsList   []Hunt
	huntsMap    map[string]Hunt
	pokedex     PokedexData
	lootPrices  map[string]int
)

// Itens protegidos que NUNCA devem ser vendidos no auto-sell
var ProtectedItems = map[string]bool{
	"pokeball":             true,
	"greatball":            true,
	"superball":            true,
	"ultraball":            true,
	"premierball":          true,
	"solariball":           true,
	"potion":               true,
	"great_potion":         true,
	"ultra_potion":         true,
	"hyper_potion":         true,
	"full_restore":         true,
	"revive":               true,
	"max_revive":           true,
	"medicine":             true,
	"candy":                true,
	"deluxe_candy":         true,
	"devoted_token":        true,
	"pvp_token":            true,
	"leaf_stone":           true,
	"fire_stone":           true,
	"water_stone":          true,
	"venom_stone":          true,
	"thunder_stone":        true,
	"rock_stone":           true,
	"earth_stone":          true,
	"coccon_stone":         true,
	"crystal_stone":        true,
	"darkness_stone":       true,
	"enigma_stone":         true,
	"heart_stone":          true,
	"ice_stone":            true,
	"metal_stone":          true,
	"feather_stone":        true,
	"held_ticket_t3":       true,
	"held_ticket_t4":       true,
	"held_ticket_t5":       true,
	"held_ticket_t6":       true,
	"held_ticket_t7":       true,
	"boost_stone":          true,
	"rough_gemstone":       true,
	"elixir_of_knowledge":  true,
	"fortune_elixir":       true,
	"abobora_halloween":    true,
}

// InitCatalog inicializa o banco de dados em memória a partir dos JSONs embutidos
func InitCatalog() {
	catalogOnce.Do(func() {
		huntsMap = make(map[string]Hunt)
		if err := json.Unmarshal(huntsJSON, &huntsList); err != nil {
			log.Printf("[Catálogo] Erro ao carregar hunts.json: %v", err)
		} else {
			for _, h := range huntsList {
				huntsMap[h.Slug] = h
			}
			log.Printf("[Catálogo] %d áreas de caça (hunts) carregadas.", len(huntsList))
		}

		if err := json.Unmarshal(pokedexJSON, &pokedex); err != nil {
			log.Printf("[Catálogo] Erro ao carregar pokedex.json: %v", err)
		} else {
			log.Printf("[Catálogo] %d Pokémon e tabela de tipos carregados.", len(pokedex.Pokemon))
		}

		lootPrices = make(map[string]int)
		var ld LootData
		if err := json.Unmarshal(lootJSON, &ld); err != nil {
			log.Printf("[Catálogo] Erro ao carregar loot.json: %v", err)
		} else {
			for k, item := range ld.Itens {
				if item.Venda > 0 {
					lootPrices[k] = item.Venda
				}
			}
			log.Printf("[Catálogo] %d preços de loot e itens registrados.", len(lootPrices))
		}
	})
}

// GetEffectiveness calcula o multiplicador de dano elemental (fraqueza/resistência)
func GetEffectiveness(moveType string, targetTypes []string) float64 {
	InitCatalog()
	row, ok := pokedex.Tipos[strings.ToLower(moveType)]
	if !ok {
		return 1.0
	}
	mult := 1.0
	for _, t := range targetTypes {
		if val, has := row[strings.ToLower(t)]; has {
			mult *= val
		}
	}
	return mult
}

// GetPokemonEntry devolve os dados da Pokédex para uma espécie
func GetPokemonEntry(especie string) (PokedexEntry, bool) {
	InitCatalog()
	entry, ok := pokedex.Pokemon[especie]
	return entry, ok
}

// GetHunt devolve a área pelo slug
func GetHunt(slug string) (Hunt, bool) {
	InitCatalog()
	h, ok := huntsMap[slug]
	return h, ok
}

// GetAllHunts devolve todas as hunts carregadas
func GetAllHunts() []Hunt {
	InitCatalog()
	return huntsList
}

// FindBestHunt calcula a rota ótima de caça considerando o nível do treinador, progresso do PokéLog e fraquezas elementais
func FindBestHunt(team []PokemonState, trainerLevel int, dex map[string]DexEntry) (Hunt, string) {
	InitCatalog()
	if len(huntsList) == 0 {
		return Hunt{}, "nenhuma área carregada"
	}

	// 1. Identificar o Pokémon ativo e seus tipos e golpes
	var primaryPoke PokemonState
	for _, p := range team {
		if p.TeamSlot != nil && *p.TeamSlot == 0 {
			primaryPoke = p
			break
		}
	}
	if primaryPoke.Especie == "" && len(team) > 0 {
		primaryPoke = team[0]
	}

	pokeEntry, hasEntry := GetPokemonEntry(primaryPoke.Especie)
	var pokeMoveTypes []string
	if hasEntry {
		for _, m := range pokeEntry.Moves {
			if !m.Passiva && m.Power > 0 {
				pokeMoveTypes = append(pokeMoveTypes, m.Tipo)
			}
		}
		if len(pokeMoveTypes) == 0 {
			pokeMoveTypes = pokeEntry.Tipos
		}
	} else {
		pokeMoveTypes = []string{"normal"}
	}

	var bestHunt Hunt
	var bestScore float64 = -1.0
	var reason string

	for _, h := range huntsList {
		if len(h.Nivel) < 2 {
			continue
		}
		minNivel := h.Nivel[0]
		maxNivel := h.Nivel[1]

		// Regra de segurança: não ir a áreas acima do nível do treinador
		if minNivel > trainerLevel {
			continue
		}

		// Penalidade para áreas muito abaixo do nível quando for farm comum
		levelDiff := trainerLevel - maxNivel
		levelFactor := 1.0
		if levelDiff > 25 {
			levelFactor = 0.6
		} else if levelDiff > 15 {
			levelFactor = 0.8
		}

		// Calcular vantagem elemental contra os monstros da rota
		var totalEffectiveness float64 = 0
		var monsterCount float64 = 0
		var incompletePokelog float64 = 0

		for _, monsterName := range h.Pokes {
			mEntry, exists := GetPokemonEntry(monsterName)
			if !exists {
				continue
			}

			// Achar o melhor golpe do nosso Pokémon contra esse monstro
			var maxEff float64 = 0.5
			for _, mType := range pokeMoveTypes {
				eff := GetEffectiveness(mType, mEntry.Tipos)
				if eff > maxEff {
					maxEff = eff
				}
			}
			totalEffectiveness += maxEff
			monsterCount++

			// Checagem de Bestiário (PokéLog):
			// Se o jogador ainda não coletou os 3 estágios dessa espécie, o monstro é prioritário
			if dex != nil {
				dEntry, inDex := dex[monsterName]
				if !inDex || dEntry.Coletado < 3 {
					estagiosFaltando := 3
					if inDex {
						estagiosFaltando = 3 - dEntry.Coletado
						if estagiosFaltando < 1 {
							estagiosFaltando = 1
						}
					}
					incompletePokelog += float64(estagiosFaltando)
				}
			}
		}

		avgEffectiveness := 1.0
		if monsterCount > 0 {
			avgEffectiveness = totalEffectiveness / monsterCount
		}

		// Bônus de Bestiário (PokéLog): rotas com espécies incompletas ganham até 3.5x de bônus!
		bestiaryFactor := 1.0
		if monsterCount > 0 && incompletePokelog > 0 {
			bestiaryFactor = 1.0 + (incompletePokelog / monsterCount) * 1.5
		}

		// Se a rota tem PokéLog incompleto, reduz a penalidade de nível baixo pois o abate rápido compensa
		if incompletePokelog > 0 && levelFactor < 1.0 {
			levelFactor = (levelFactor + 1.0) / 2.0
		}

		avgLevel := float64(minNivel+maxNivel) / 2.0
		// Score = Kills/h * Vantagem Elemental * Bestiário * Nível
		score := float64(h.KillsPorHora) * avgEffectiveness * bestiaryFactor * levelFactor * (1.0 + (avgLevel * 0.03))

		if score > bestScore {
			bestScore = score
			bestHunt = h
			pokelogMsg := ""
			if incompletePokelog > 0 {
				pokelogMsg = fmt.Sprintf(" | 📖 Bestiário Pendente (%.0f metas)", incompletePokelog)
			}
			reason = fmt.Sprintf("Lv %d-%d | Eficácia Elemental: %.1fx%s | Score: %.0f", minNivel, maxNivel, avgEffectiveness, pokelogMsg, score)
		}
	}

	if bestHunt.Slug == "" && len(huntsList) > 0 {
		return huntsList[0], "área inicial padrão"
	}

	return bestHunt, reason
}

// PickBestAttacker escolhe qual pokémon do time tem a melhor vantagem de tipo contra uma lista de espécies respeitando obediência
func PickBestAttacker(team []PokemonState, huntMonsters []string, trainerLevel int) int {
	InitCatalog()
	if len(team) == 0 {
		return 0
	}

	var bestID = team[0].ID
	var highestAdv float64 = -1.0

	for _, p := range team {
		pEntry, ok := GetPokemonEntry(p.Especie)
		if !ok {
			continue
		}

		// Validação crítica de obediência do PokeTrue:
		// Se o nível de treinador for menor que o requisito da espécie, o Pokémon não obedece e fica travado!
		if trainerLevel > 0 && pEntry.NivelReq > trainerLevel {
			continue
		}

		var totalScore float64 = 0
		for _, mName := range huntMonsters {
			mEntry, mOk := GetPokemonEntry(mName)
			if !mOk {
				continue
			}

			// Dano causado pelo nosso monstro
			var bestAttack float64 = 0.5
			for _, m := range pEntry.Moves {
				if m.Power > 0 {
					eff := GetEffectiveness(m.Tipo, mEntry.Tipos)
					if eff > bestAttack {
						bestAttack = eff
					}
				}
			}

			// Dano recebido do monstro (risco defensivo)
			var worstDefense float64 = 1.0
			for _, m := range mEntry.Moves {
				if m.Power > 0 {
					eff := GetEffectiveness(m.Tipo, pEntry.Tipos)
					if eff > worstDefense {
						worstDefense = eff
					}
				}
			}

			// Vantagem relativa
			totalScore += (bestAttack / worstDefense)
		}

		if totalScore > highestAdv {
			highestAdv = totalScore
			bestID = p.ID
		}
	}

	return bestID
}

// GetSellableLoot filtra o inventário e devolve itens seguros para venda em lote
func GetSellableLoot(stash map[string]int) []SellItem {
	InitCatalog()
	var list []SellItem
	for item, count := range stash {
		if count <= 0 {
			continue
		}
		if ProtectedItems[item] {
			continue
		}
		// Apenas itens cadastrados no loot com valor de venda
		if price, has := lootPrices[item]; has && price > 0 {
			list = append(list, SellItem{
				Item: item,
				Qtd:  count,
			})
		}
	}
	return list
}

// GetBestPotion identifica a melhor poção para a quantidade de vida perdida
func GetBestPotion(missingHP int, stash map[string]int) string {
	potions := []struct {
		name string
		heal int
	}{
		{"hyper_potion", 1000},
		{"ultra_potion", 400},
		{"great_potion", 150},
		{"potion", 60},
	}

	for _, pot := range potions {
		if (stash[pot.name] > 0) && (missingHP >= pot.heal/2) {
			return pot.name
		}
	}

	// Fallback para qualquer poção existente
	for _, pot := range potions {
		if stash[pot.name] > 0 {
			return pot.name
		}
	}

	if stash["full_restore"] > 0 && missingHP > 1000 {
		return "full_restore"
	}

	return ""
}
