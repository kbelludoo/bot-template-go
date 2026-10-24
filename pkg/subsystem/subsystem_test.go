package subsystem

import (
	"testing"
	"time"
)

func TestRouteSolver(t *testing.T) {
	areas := []AreaCandidate{
		{
			ID:           "larva_woods",
			Name:         "Mata das Larvas",
			MinLevel:     1,
			MaxLevel:     15,
			BaseExp:      50,
			KillsPerHour: 400,
			Monsters:     []string{"Caterpie", "Weedle"},
		},
		{
			ID:           "goblin_cave",
			Name:         "Caverna dos Goblins",
			MinLevel:     15,
			MaxLevel:     30,
			BaseExp:      150,
			KillsPerHour: 350,
			Monsters:     []string{"Goblin", "Goblin Scout"},
		},
		{
			ID:           "dragon_lair",
			Name:         "Cova dos Dragões",
			MinLevel:     50,
			MaxLevel:     80,
			BaseExp:      800,
			KillsPerHour: 200,
			Monsters:     []string{"Dragon", "Dragon Lord"},
		},
	}

	solver := NewRouteSolver(areas)

	// Treinador Nível 20 com ataque de Fogo contra Goblins
	crit := SolverCriteria{
		CharacterLevel:    20,
		PrimaryAttackType: "fire",
		TargetBestiary: map[string]bool{
			"Goblin": true, // Bestiário pendente
		},
		ElementalMatrix: func(atk, mon string) float64 {
			if mon == "Goblin" {
				return 1.5 // 1.5x de fraqueza a fogo
			}
			return 1.0
		},
	}

	best, ok := solver.FindOptimalRoute(crit)
	if !ok {
		t.Fatalf("Solver não retornou rota")
	}

	if best.Area.ID != "goblin_cave" {
		t.Errorf("Esperava 'goblin_cave', obteve '%s'", best.Area.ID)
	}
	if !best.IsBestiary {
		t.Errorf("Deveria detectar bestiário pendente na rota")
	}
	t.Logf("Melhor rota: %s (Score: %.0f, Motivo: %s)", best.Area.Name, best.Score, best.Reason)
}

func TestInventoryManager_AutoSell(t *testing.T) {
	im := NewInventoryManager(100, []string{"fire_stone", "potion", "rare_candy"})
	im.SetPrice("caterpie_fur", 15)
	im.SetPrice("seed", 10)
	im.SetPrice("gold_nugget", 500)

	stash := map[string]int{
		"fire_stone":   2,   // Protegido
		"caterpie_fur": 50,  // Venda: 50 * 15 = 750
		"seed":         100, // Venda: 100 * 10 = 1000
		"potion":       20,  // Protegido
	}

	batch := im.GetSellableBatch(stash)
	if len(batch.Items) != 2 {
		t.Fatalf("Esperava 2 itens no lote de venda, obteve: %d", len(batch.Items))
	}
	if batch.TotalGold != 1750 {
		t.Errorf("Total de ouro esperado 1750, obteve: %d", batch.TotalGold)
	}

	// Testar verificação de capacidade
	if im.ShouldSell(50, 0.8) {
		t.Errorf("50/100 não deveria disparar venda a 80%%")
	}
	if !im.ShouldSell(85, 0.8) {
		t.Errorf("85/100 deveria disparar venda a 80%%")
	}
}

func TestDailyManager(t *testing.T) {
	dm := NewDailyManager(time.UTC)

	// Inicialmente precisa resgatar
	if !dm.CanClaimRewards() {
		t.Errorf("Deveria permitir resgate inicial")
	}

	dm.MarkClaimed()
	if dm.CanClaimRewards() {
		t.Errorf("Não deveria permitir novo resgate no mesmo dia")
	}
	if dm.GetStreak() != 1 {
		t.Errorf("Streak esperado 1, obteve: %d", dm.GetStreak())
	}
}

func TestPotionManager(t *testing.T) {
	potions := []PotionDefinition{
		{ID: "small_pot", Name: "Poção Pequena", HealPower: 50, CostGold: 10},
		{ID: "medium_pot", Name: "Poção Média", HealPower: 150, CostGold: 30},
		{ID: "big_pot", Name: "Poção Grande", HealPower: 400, CostGold: 90},
	}

	pm := NewPotionManager(potions, 10)
	stash := map[string]int{
		"small_pot":  20,
		"medium_pot": 15,
		"big_pot":    5,
	}

	// Perdeu 80 de HP -> medium_pot (150/2 = 75) é indicada
	p := pm.SelectBestHPPotion(80, stash)
	if p != "medium_pot" {
		t.Errorf("Esperava medium_pot para perda de 80 HP, obteve: %s", p)
	}

	// Perdeu 250 de HP -> big_pot
	p2 := pm.SelectBestHPPotion(250, stash)
	if p2 != "big_pot" {
		t.Errorf("Esperava big_pot para perda de 250 HP, obteve: %s", p2)
	}

	// Checar restock: big_pot tem 5 (< 10)
	toBuy := pm.NeedsRestock(stash)
	if len(toBuy) != 1 || toBuy[0] != "big_pot" {
		t.Errorf("Esperava big_pot para recompra, obteve: %v", toBuy)
	}
}

func TestPromotionManager(t *testing.T) {
	tiers := []PromotionTier{
		{LevelReq: 20, CostGold: 20000, TargetName: "Elite Knight", ActionCode: "promote_1"},
		{LevelReq: 100, CostGold: 500000, TargetName: "Exalted Knight", ActionCode: "promote_2"},
	}
	bCfg := BlessingConfig{
		MaxBlessings: 5,
		CostPerBless: 10000,
	}

	pm := NewPromotionManager(tiers, bCfg)

	// Nível 25 com 25.000 de ouro e classe "Knight" -> Elegível
	tier, ok := pm.CheckEligiblePromotion(25, 25000, "Knight")
	if !ok || tier.TargetName != "Elite Knight" {
		t.Errorf("Esperava promoção para Elite Knight, obteve: %+v", tier)
	}

	// Já é "Elite Knight" -> Não elegível para o mesmo tier
	_, ok2 := pm.CheckEligiblePromotion(25, 25000, "Elite Knight")
	if ok2 {
		t.Errorf("Não deveria permitir promoção repetida para Elite Knight")
	}

	// Bênçãos
	missing := pm.MissingBlessings(3)
	if missing != 2 {
		t.Errorf("Esperava 2 bênçãos faltantes, obteve: %d", missing)
	}
	if !pm.CanAffordBlessings(3, 25000) {
		t.Errorf("Deveria poder comprar 2 bênçãos com 25k (custo: 20k)")
	}
}

func TestPartyCoordinator(t *testing.T) {
	coord := NewPartyCoordinator("party-1", "acc-1", 40.0, 0.45)

	coord.RegisterMember(PartyMember{ID: "acc-1", Name: "Tanker", Level: 100})
	coord.RegisterMember(PartyMember{ID: "acc-2", Name: "DPS 1", Level: 105})
	coord.RegisterMember(PartyMember{ID: "acc-3", Name: "DPS 2", Level: 102})
	coord.RegisterMember(PartyMember{ID: "acc-4", Name: "Healer", Level: 98})

	if !coord.IsLeader("acc-1") {
		t.Errorf("acc-1 deveria ser o líder")
	}

	// Com 4 membros e bônus de 45%:
	// KPM individual <= 40 / (4 * 1.45) = 40 / 5.8 ≈ 6.896 kills/min
	kpm := coord.CalculateThrottledKillRate()
	if kpm < 6.8 || kpm > 7.0 {
		t.Errorf("KPM de grupo incorreto: %f (esperado ~6.89)", kpm)
	}

	interval := coord.GetSafetyIntervalSec()
	if interval < 8.5 || interval > 8.9 {
		t.Errorf("Intervalo de segurança incorreto: %f seg (esperado ~8.7s)", interval)
	}
	t.Logf("Coordenador de Grupo: %s", coord.String())
}

func TestBossScheduler(t *testing.T) {
	bs := NewBossScheduler()
	now := time.Now()

	bs.AddBoss(WorldBossEvent{
		ID:        "ferumbras",
		Name:      "Ferumbras",
		SpawnTime: now.Add(-2 * time.Minute), // Spawnou há 2 minutos
		MinLevel:  100,
		IsOpen:    true,
	})

	bs.AddBoss(WorldBossEvent{
		ID:        "morgaroth",
		Name:      "Morgaroth",
		SpawnTime: now.Add(2 * time.Hour), // Futuro
		MinLevel:  150,
		IsOpen:    false,
	})

	ready := bs.GetReadyBosses(120)
	if len(ready) != 1 || ready[0].ID != "ferumbras" {
		t.Errorf("Esperava somente Ferumbras pronto, obteve: %+v", ready)
	}

	// Após marcar como concluído
	bs.MarkCompleted("ferumbras")
	readyAfter := bs.GetReadyBosses(120)
	if len(readyAfter) != 0 {
		t.Errorf("Não deveria retornar chefes após marcar completado")
	}
}

func TestCombatSimAndPokemonMatrix(t *testing.T) {
	sim := SimulateCombat(CombatSimInput{
		PlayerHP:             1000,
		PlayerATK:            150,
		PlayerDEF:            40,
		PlayerAtkIntervalSec: 1.0,
		PlayerHealPerSec:     30,
		MobHP:                300,
		MobATK:               55,
		MobDEF:               20,
		MobAtkIntervalSec:    1.5,
		MobDensity:           2,
	})
	if !sim.CanTank || sim.TTKSec <= 0 {
		t.Fatalf("esperava tankar com folga: %+v", sim)
	}

	if mult := PokemonElementalEffectiveness("fire", "grass"); mult != 2.0 {
		t.Fatalf("esperava 2.0x fire->grass, obteve %v", mult)
	}
	if mult := PokemonElementalEffectiveness("electric", "ground"); mult != 0.0 {
		t.Fatalf("esperava 0.0x electric->ground, obteve %v", mult)
	}
}

func TestPartyLevelSyncAndLurker(t *testing.T) {
	coord := NewPartyCoordinator("p2", "a1", 40, 0.45)
	coord.RegisterMember(PartyMember{ID: "a1", Level: 50})
	coord.RegisterMember(PartyMember{ID: "a2", Level: 68})

	if coord.MinMemberLevel() != 50 {
		t.Fatalf("esperava MinMemberLevel=50, obteve %d", coord.MinMemberLevel())
	}
	if !coord.CanFormPartyWithLevelGap(25) {
		t.Fatalf("diferença de 18 níveis deveria ser aceita para maxGap=25")
	}

	coord.RecordMemberBestiary("a1", "orc", 500)
	coord.RecordMemberBestiary("a2", "orc", 120)
	if shared := coord.GetSharedBestiaryKills("orc"); shared != 120 {
		t.Fatalf("esperava shared kills = 120, obteve %d", shared)
	}

	coord.LockHunt("abyss_gate", 5*time.Minute)
	if !coord.IsHuntLocked("abyss_gate") {
		t.Fatalf("abyss_gate deveria estar bloqueada")
	}

	joins := FormatTwitchIRCJoin([]string{"#StreamerA", "StreamerB"})
	if len(joins) != 2 || joins[0] != "JOIN #streamera" {
		t.Fatalf("joins inesperados: %v", joins)
	}
}
