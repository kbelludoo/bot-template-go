package game

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"bot-template-go/pkg/engine"
)

// Bot implementa o bot autônomo completo para PokeTrue
type Bot struct {
	config  AccountConfig
	mu      sync.RWMutex
	status  engine.AccountStatus
	api     *PokeTrueClient
	cb      *engine.CircuitBreaker

	// Estado canônico do jogo
	state        *PokeTrueState
	currentHunt  Hunt
	activePokeID int
	serverCombat bool

	// Controles temporais
	startTime       time.Time
	initialXp       int
	lastSyncAt      time.Time
	lastSellAt      time.Time
	lastHeldAt      time.Time
	lastTreinoAt    time.Time
	lastJennyAt     time.Time
	lastCandyAt     time.Time
	lastHuntChangeAt time.Time
	lastPokeLogAt   time.Time
	lastRunasAt     time.Time
	lastEvoAt       time.Time
	lastGinasioAt   time.Time
	lastBossAt      time.Time
	lastVipCheckAt  time.Time
	lastSwitchedPokeID int

	// Canais para supply externo de tickets/tokens
	ticketChan    chan string
	turnstileChan chan string
}

// NewBot inicializa uma nova instância do bot do PokeTrue
func NewBot(cfg AccountConfig) *Bot {
	InitCatalog()

	cbCfg := engine.DefaultCircuitBreakerConfig()
	cb := engine.NewCircuitBreaker(cfg.ID, cbCfg)

	st := engine.NewStatus(cfg.ID, cfg.Name)
	st.CurrentArea = cfg.Settings.HuntTarget
	if st.CurrentArea == "" {
		st.CurrentArea = "auto"
	}
	st.Status = "Iniciando"

	client := NewPokeTrueClient(cfg.Credentials.Token, cfg.BindIP, cb)

	b := &Bot{
		config:        cfg,
		status:        st,
		api:           client,
		cb:            cb,
		ticketChan:    make(chan string, 1),
		turnstileChan: make(chan string, 1),
	}

	cbCfg.OnTrip = func(reason string) {
		b.addLog(fmt.Sprintf("⚠️ Anti-Cheat disparado: %s. Pausando combate e recuando.", reason))
		b.updateStatus(func(s *engine.AccountStatus) {
			s.Status = "Proteção Anti-Cheat Ativa (Cooldown)"
		})
		_ = b.api.StopHunt(context.Background())
	}

	cbCfg.OnReset = func() {
		b.addLog("✅ Cooldown concluído. Retomando caça normal.")
		b.updateStatus(func(s *engine.AccountStatus) {
			s.Status = "hunting"
		})
	}

	return b
}

func (b *Bot) ID() string {
	return b.config.ID
}

func (b *Bot) Name() string {
	return b.config.Name
}

func (b *Bot) Init(ctx context.Context) error {
	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "Autenticando..."
	})

	// 1. Validar se já possui token JWT ou se precisa autenticar via login
	if b.api.GetToken() == "" {
		if b.config.Credentials.Email != "" && b.config.Credentials.Password != "" {
			b.addLog("Autenticando com credenciais de login...")
			resp, err := b.api.Login(ctx, b.config.Credentials.Email, b.config.Credentials.Password, "")
			if err != nil {
				b.addLog(fmt.Sprintf("Aviso: Falha no login direto (%v). Aguardando token...", err))
			} else if resp.Token != "" {
				b.api.SetToken(resp.Token)
				b.addLog("Login realizado com sucesso!")
			}
		}
	}

	if b.api.GetToken() == "" {
		b.updateStatus(func(s *engine.AccountStatus) {
			s.Status = "Aguardando Token"
		})
		b.addLog("Atenção: Nenhum token de sessão configurado. Copie o token de localStorage['pokeidle.token'] ou rode token-grabber.")
		return fmt.Errorf("token de sessão ausente")
	}

	// 2. Carregar perfil (me) e estado do jogo (state)
	me, err := b.api.GetMe(ctx)
	if err != nil {
		b.addLog(fmt.Sprintf("Erro ao consultar /api/me: %v", err))
	} else {
		b.addLog(fmt.Sprintf("Treinador '%s' conectado com sucesso (Vip: %v, Diamantes: %d).", me.User, me.Vip, me.Diamonds))
		if len(me.Personagens) > 0 {
			b.addLog(fmt.Sprintf("Entrando com personagem '%s' (#%d)...", me.Personagens[0].Nome, me.Personagens[0].ID))
			_, _ = b.api.EscolherPersonagem(ctx, me.Personagens[0].ID)
		}
	}

	st, err := b.api.GetState(ctx)
	if err != nil {
		return fmt.Errorf("falha ao carregar estado inicial do jogo: %w", err)
	}

	b.mu.Lock()
	b.state = st
	b.initialXp = st.NivelXp
	b.startTime = time.Now()
	b.mu.Unlock()

	b.refreshStatusFromState(st)
	b.addLog(fmt.Sprintf("Estado carregado: Treinador Nível %d | Ouro: %d | Time: %d Pokémon.", st.Nivel, st.Gold, len(st.Pokemon)))
	return nil
}

func (b *Bot) Run(ctx context.Context) error {
	b.updateStatus(func(s *engine.AccountStatus) {
		s.Status = "hunting"
	})
	b.addLog("⚡ Ciclo de caça autônomo iniciado.")

	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.addLog("Encerrando bot...")
			_ = b.api.StopHunt(context.Background())
			return ctx.Err()

		case tk := <-b.ticketChan:
			b.api.SetToken(tk)
			b.addLog("Novo token de sessão aplicado.")
			_ = b.Init(ctx)

		case <-ticker.C:
			if !b.cb.AllowAction() {
				continue
			}

			if err := b.performBotStep(ctx); err != nil {
				b.addLog(fmt.Sprintf("Aviso no ciclo: %v", err))
				time.Sleep(2 * time.Second)
			}
		}
	}
}

// performBotStep executa um ciclo completo de caça, fraquezas, cura, helds, treino e loot
func (b *Bot) performBotStep(ctx context.Context) error {
	b.mu.RLock()
	st := b.state
	b.mu.RUnlock()
	if st == nil || b.api.GetToken() == "" {
		return nil
	}

	now := time.Now()

	// 1. Sincronização periódica do progresso da caçada
	syncInterval := 20 * time.Second
	if b.serverCombat {
		syncInterval = 5 * time.Second
	}

	if now.Sub(b.lastSyncAt) > syncInterval {
		b.lastSyncAt = now
		activeHp := b.getActivePokemonHP()
		syncResp, err := b.api.SyncHunt(ctx, activeHp, false)
		if err == nil && syncResp != nil {
			if syncResp.Delta != nil {
				b.mu.Lock()
				b.status.Kills += syncResp.Delta.Kills
				b.status.Gold += syncResp.Delta.Gold
				if syncResp.Delta.Treinador != nil {
					b.status.Level = syncResp.Delta.Treinador.Nivel
				}
				b.mu.Unlock()

				for _, cap := range syncResp.Delta.Capturas {
					b.addLog(fmt.Sprintf("🎉 CAPTURA REALIZADA: %s (Lv %d)!", cap.Especie, cap.Nivel))
				}
			}
			if syncResp.State != nil {
				b.mu.Lock()
				b.state = syncResp.State
				b.mu.Unlock()
				b.refreshStatusFromState(syncResp.State)
			}
		}
	}

	// 2. Resolver Rota Ótima (Fraquezas Elementais & Melhores Áreas)
	if err := b.manageOptimalHunt(ctx); err != nil {
		return err
	}

	// 3. Processar Fila de Combate (apenas se combate local ativo)
	if !b.serverCombat {
		if err := b.processCombatQueue(ctx); err != nil {
			return err
		}
	}

	// 4. Gestão de Vida e Auto-Heal (Stamina de HP)
	if b.config.Settings.AutoHeal {
		b.manageAutoHeal(ctx)
	}

	// 5. Gestão de Helds / Equips (Auto-Equip e Fusão de Tiers)
	if b.config.Settings.AutoEquip && now.Sub(b.lastHeldAt) > 5*time.Minute {
		b.lastHeldAt = now
		go b.manageHelds(context.Background())
	}

	// 6. Gestão de Campo de Treino (Treino de EVs dos Pokémon de reserva)
	if b.config.Settings.AutoTreino && now.Sub(b.lastTreinoAt) > 4*time.Minute {
		b.lastTreinoAt = now
		go b.manageTreino(context.Background())
	}

	// 7. Auto-Sell inteligente de Loot acumulado
	if b.config.Settings.AutoSell && now.Sub(b.lastSellAt) > 3*time.Minute {
		b.lastSellAt = now
		go b.manageAutoSell(context.Background())
	}

	// 8. Missões da Oficial Jenny & Tarefas Diárias
	if b.config.Settings.AutoDaily && now.Sub(b.lastJennyAt) > 8*time.Minute {
		b.lastJennyAt = now
		go b.manageJenny(context.Background())
	}

	// 9. Consumo automático de Rare Candies para acelerar evolução
	if now.Sub(b.lastCandyAt) > 5*time.Minute {
		b.lastCandyAt = now
		go b.manageCandies(context.Background())
	}

	// 10. Gestão de PokéLog / Bestiário (Resgate automático de marcos de abates alcançados)
	if now.Sub(b.lastPokeLogAt) > 2*time.Minute {
		b.lastPokeLogAt = now
		go b.managePokeLog(context.Background())
	}

	// 11. Gestão de Runas do PokéLog (Compra e ativação de bônus passivos de EXP e Loot)
	if now.Sub(b.lastRunasAt) > 6*time.Minute {
		b.lastRunasAt = now
		go b.manageRunas(context.Background())
	}

	// 12. Gestão de Evolução de Pokémon (Nível e pedras elementais)
	if now.Sub(b.lastEvoAt) > 5*time.Minute {
		b.lastEvoAt = now
		go b.manageEvolucao(context.Background())
	}

	// 13. Gestão de Ginásios & Insígnias (Bônus permanente de dano)
	if now.Sub(b.lastGinasioAt) > 10*time.Minute {
		b.lastGinasioAt = now
		go b.manageGinasios(context.Background())
	}

	// 14. Gestão de World Boss diário (Entrada grátis diária)
	if now.Sub(b.lastBossAt) > 15*time.Minute {
		b.lastBossAt = now
		go b.manageBoss(context.Background())
	}

	// 15. Gestão de VIP Automática (Fase 1: 30 dias por 25 diamantes | Fase 2: 180 dias por 135 diamantes)
	if now.Sub(b.lastVipCheckAt) > 5*time.Minute {
		b.lastVipCheckAt = now
		go b.manageAutoVIP(context.Background())
	}

	return nil
}

// manageOptimalHunt calcula a melhor área considerando fraquezas elementais e metas do Bestiário
func (b *Bot) manageOptimalHunt(ctx context.Context) error {
	b.mu.RLock()
	st := b.state
	targetArea := b.config.Settings.HuntTarget
	b.mu.RUnlock()

	if st == nil {
		return nil
	}

	var desiredHunt Hunt
	var reason string

	if targetArea == "" || targetArea == "auto" {
		bestHunt, why := FindBestHunt(st.Pokemon, st.Nivel, st.Dex)
		desiredHunt = bestHunt
		reason = why
	} else {
		h, exists := GetHunt(targetArea)
		if exists {
			desiredHunt = h
			reason = "área configurada manualmente"
		} else {
			bestHunt, why := FindBestHunt(st.Pokemon, st.Nivel, st.Dex)
			desiredHunt = bestHunt
			reason = why
		}
	}

	if desiredHunt.Slug == "" {
		return nil
	}

	// Se a caça atual no servidor for diferente da desejada, transita para a nova área
	if st.Hunt.Mapa != desiredHunt.Slug {
		// Evitar trocas frequentes de rota (rate limit do servidor): no máximo 1 troca a cada 10 minutos
		if st.Hunt.Mapa != "" && time.Since(b.lastHuntChangeAt) < 10*time.Minute && b.currentHunt.Slug != "" {
			return nil
		}
		b.lastHuntChangeAt = time.Now()
		b.addLog(fmt.Sprintf("🗺️ Transicionando para rota ótima: '%s' (%s)", desiredHunt.Nome, reason))
		startResp, err := b.api.StartHunt(ctx, desiredHunt.Slug, map[string]any{}, false)
		if err != nil {
			return fmt.Errorf("erro iniciando hunt %s: %w", desiredHunt.Slug, err)
		}
		if startResp.State != nil {
			b.mu.Lock()
			b.state = startResp.State
			b.currentHunt = desiredHunt
			b.mu.Unlock()
			b.refreshStatusFromState(startResp.State)
		}
	}

	// Escolher o melhor Pokémon atacante do time para essa rota respeitando obediência
	bestPokeID := PickBestAttacker(st.Pokemon, desiredHunt.Pokes, st.Nivel)
	if bestPokeID > 0 && bestPokeID != st.Hunt.AtivoID && bestPokeID != b.lastSwitchedPokeID {
		b.addLog(fmt.Sprintf("🔄 Alternando atacante ativo para Pokémon #%d (maior vantagem elemental).", bestPokeID))
		_ = b.api.SwitchActive(ctx, bestPokeID)
		b.lastSwitchedPokeID = bestPokeID
	}

	return nil
}

// processCombatQueue abate monstros na fila respeitando pacing anti-ban e captura
func (b *Bot) processCombatQueue(ctx context.Context) error {
	b.mu.RLock()
	st := b.state
	b.mu.RUnlock()

	if st == nil || len(st.Hunt.Fila) == 0 {
		return nil
	}

	for _, monster := range st.Hunt.Fila {
		if monster.Morto {
			continue
		}

		// 1. Avaliar auto-captura (prioriza shinies ou bichos configurados)
		if b.config.Settings.AutoCatch || monster.Shiny {
			b.attemptCatch(ctx, monster)
		}

		// 2. Realizar abate
		activeHpRatio := 1.0
		activeFicha := b.getActivePokemon()
		if activeFicha != nil && activeFicha.HpMax > 0 {
			activeHpRatio = math.Max(0.1, activeFicha.Hp/float64(activeFicha.HpMax))
		}

		killResp, err := b.api.KillMonster(ctx, monster.ID, activeHpRatio)
		if err != nil {
			if strings.Contains(err.Error(), "combate_no_servidor") {
				b.serverCombat = true
				b.addLog("⚡ Combate automático ativo no servidor (idle nativo). Sincronizando progresso...")
				return nil
			}
			b.addLog(fmt.Sprintf("Erro ao abater monstro %s #%s: %v", monster.Especie, monster.ID, err))
			return nil
		}

		b.mu.Lock()
		b.status.Kills++
		if killResp.Delta != nil {
			b.status.Gold += killResp.Delta.Gold
			if killResp.Delta.Treinador != nil {
				b.status.Level = killResp.Delta.Treinador.Nivel
			}
		}
		b.mu.Unlock()

		if killResp.Capturado != nil {
			b.addLog(fmt.Sprintf("🎉 CAPTURA REALIZADA: %s (Lv %d%s)!",
				killResp.Capturado.Especie, killResp.Capturado.Nivel,
				func() string {
					if killResp.Capturado.Shiny {
						return " SHINY ✨"
					}
					return ""
				}()))
		}

		// Ritmo humano realista entre abates (300ms a 700ms)
		time.Sleep(350 * time.Millisecond)
		break // Processa um monstro por ciclo do loop
	}

	return nil
}

// attemptCatch tenta arremessar a pokébola adequada
func (b *Bot) attemptCatch(ctx context.Context, m FilaMonster) {
	b.mu.RLock()
	stash := b.state.Itens
	b.mu.RUnlock()

	ball := "pokeball"
	if m.Shiny {
		for _, highBall := range []string{"ultraball", "solariball", "premierball", "superball", "greatball"} {
			if stash[highBall] > 0 {
				ball = highBall
				break
			}
		}
	} else if b.config.Settings.BallType != "" && b.config.Settings.BallType != "auto" {
		ball = b.config.Settings.BallType
	} else {
		for _, bType := range []string{"pokeball", "greatball", "superball"} {
			if stash[bType] > 0 {
				ball = bType
				break
			}
		}
	}

	if stash[ball] > 0 {
		_, _ = b.api.ThrowBall(ctx, ball, m.ID)
	}
}

// manageAutoHeal cuida da saúde e recuperação do Pokémon ativo
func (b *Bot) manageAutoHeal(ctx context.Context) {
	poke := b.getActivePokemon()
	if poke == nil || poke.HpMax <= 0 {
		return
	}

	threshold := b.config.Settings.HealThreshold
	if threshold <= 0 {
		threshold = 0.50
	}

	currentRatio := poke.Hp / float64(poke.HpMax)
	if currentRatio >= threshold {
		return
	}

	b.mu.RLock()
	if b.state == nil {
		b.mu.RUnlock()
		return
	}
	stash := b.state.Itens
	b.mu.RUnlock()

	missingHP := poke.HpMax - int(poke.Hp)
	potion := GetBestPotion(missingHP, stash)

	if potion != "" {
		_ = b.api.ConsumeItem(ctx, potion, 1)
		b.addLog(fmt.Sprintf("❤️ Auto-Heal: Monstro com %.0f%% HP. Usada poção '%s'.", currentRatio*100, potion))
	} else if b.status.Gold > 1000 {
		// Comprar lote emergencial de 10 poções
		b.addLog("🛒 Estoque de poções esgotado. Comprando 10 poções na loja...")
		_ = b.api.BuyItem(ctx, "potion", 10)
	}
}

// manageHelds verifica equips ideais e funde duplicatas para tiers superiores
func (b *Bot) manageHelds(ctx context.Context) {
	helds, err := b.api.GetHelds(ctx)
	if err != nil || helds == nil {
		return
	}

	// 1. Auto-Equip: Verificar se algum Pokémon do time está sem held
	b.mu.RLock()
	if b.state == nil {
		b.mu.RUnlock()
		return
	}
	team := b.state.Pokemon
	b.mu.RUnlock()

	for _, p := range team {
		if p.Held == nil && len(helds.Mochila) > 0 {
			// Achar o melhor held disponível na mochila
			var bestHeld HeldDetail
			var bestIdx = -1
			for idx, h := range helds.Mochila {
				if h.Tier > bestHeld.Tier {
					bestHeld = h
					bestIdx = idx
				}
			}
			if bestIdx >= 0 {
				b.addLog(fmt.Sprintf("🛡️ Auto-Equip: Equipando held '%s' Tier %d no Pokémon #%d.", bestHeld.Tipo, bestHeld.Tier, p.ID))
				_ = b.api.EquipHeld(ctx, p.ID, bestHeld.ID)
				helds.Mochila = append(helds.Mochila[:bestIdx], helds.Mochila[bestIdx+1:]...)
			}
		}
	}

	// 2. Auto-Fusão: Se houver 3 helds do mesmo tier e tipo na mochila, funde-os!
	byTierAndType := make(map[string][]int)
	for _, h := range helds.Mochila {
		if h.Tier < 6 { // Evita fundir helds raros sem critério
			key := fmt.Sprintf("%s_t%d", h.Tipo, h.Tier)
			byTierAndType[key] = append(byTierAndType[key], h.ID)
		}
	}

	for key, ids := range byTierAndType {
		if len(ids) >= 3 {
			triplet := ids[:3]
			b.addLog(fmt.Sprintf("⚡ Auto-Fusão: Fundindo 3x helds de '%s' para criar Tier superior!", key))
			_ = b.api.FuseHelds(ctx, triplet)
			break
		}
	}
}

// manageTreino gerencia o campo de treino para aumentar os atributos (EVs) de reservas
func (b *Bot) manageTreino(ctx context.Context) {
	treino, err := b.api.GetTreino(ctx, "npc")
	if err != nil || treino == nil {
		return
	}

	b.mu.RLock()
	if b.state == nil {
		b.mu.RUnlock()
		return
	}
	team := b.state.Pokemon
	gold := b.status.Gold
	b.mu.RUnlock()

	for _, slot := range treino.Slots {
		// Se a vaga estiver vazia, colocar um Pokémon que não esteja em campo para treinar Atk/SpA
		if slot.PokemonID == nil {
			for _, p := range team {
				if p.TeamSlot == nil || *p.TeamSlot != 0 {
					stat := "atk"
					pEntry, ok := GetPokemonEntry(p.Especie)
					if ok && pEntry.Base.SpA > pEntry.Base.Atk {
						stat = "spa"
					}
					b.addLog(fmt.Sprintf("🏋️ Campo de Treino: Colocando Pokémon #%d (%s) para treinar %s.", p.ID, p.Especie, stat))
					_ = b.api.SetTreino(ctx, "npc", slot.Slot, p.ID, stat)
					break
				}
			}
		}

		// Recarregar cargas de treino se estiverem abaixo de 50 e houver ouro suficiente
		if slot.Cargas < 50 && gold > 10000 {
			_ = b.api.RefillTreino(ctx, "npc", slot.Slot, "ouro")
			b.addLog(fmt.Sprintf("🔋 Campo de Treino: Recarregadas cargas da vaga %d com ouro.", slot.Slot))
		}
	}
}

// manageAutoSell vende loot acumulado sem vender itens essenciais ou protegidos
func (b *Bot) manageAutoSell(ctx context.Context) {
	b.mu.RLock()
	stash := b.state.Itens
	b.mu.RUnlock()

	itemsToSell := GetSellableLoot(stash)
	if len(itemsToSell) == 0 {
		return
	}

	totalItems := 0
	for _, it := range itemsToSell {
		totalItems += it.Qtd
	}

	b.addLog(fmt.Sprintf("💰 Auto-Sell: Vendendo %d itens de loot no NPC Mark...", totalItems))
	err := b.api.SellLote(ctx, itemsToSell)
	if err != nil {
		b.addLog(fmt.Sprintf("Aviso no auto-sell: %v", err))
	} else {
		b.addLog("✅ Venda concluída com sucesso. Ouro atualizado!")
	}
}

// manageJenny aceita e entrega missões diárias da Oficial Jenny
func (b *Bot) manageJenny(ctx context.Context) {
	jenny, err := b.api.GetJenny(ctx)
	if err == nil && jenny != nil {
		if !jenny.Escolhidas && len(jenny.Tarefas) > 0 {
			var toPick []string
			for i, t := range jenny.Tarefas {
				if i < 3 {
					toPick = append(toPick, t.Especie)
				}
			}
			if len(toPick) > 0 {
				_ = b.api.ChooseJenny(ctx, toPick)
				b.addLog(fmt.Sprintf("👮 Oficial Jenny: Missões diárias aceitas (%s).", strings.Join(toPick, ", ")))
			}
		}
	}

	tarefas, err := b.api.GetTarefas(ctx)
	if err == nil && tarefas != nil {
		for _, t := range tarefas.Tarefas {
			if t.Pronto && !t.Concluido {
				_ = b.api.CompleteTarefa(ctx, t.ID)
				b.addLog(fmt.Sprintf("🎁 Tarefa Concluída: '%s' resgatada!", t.Titulo))
			}
		}
	}
}

// manageCandies usa doces raros em Pokémon com nível abaixo de 100
func (b *Bot) manageCandies(ctx context.Context) {
	b.mu.RLock()
	if b.state == nil {
		b.mu.RUnlock()
		return
	}
	stash := b.state.Itens
	b.mu.RUnlock()

	if stash["candy"] > 0 {
		b.addLog("🍬 Consumindo Rare Candy no Pokémon ativo...")
		_ = b.api.ConsumeItem(ctx, "candy", 1)
	}
}

// managePokeLog recolhe periodicamente as metas concluídas de abates no Bestiário (PokéLog)
func (b *Bot) managePokeLog(ctx context.Context) {
	resp, err := b.api.ResgatarPokeLog(ctx, true, "")
	if err != nil {
		return
	}
	if resp != nil && len(resp.Feitos) > 0 {
		b.addLog(fmt.Sprintf("📖 POKÉLOG RESGATADO: %d marco(s) concluído(s)! (+%d XP Treinador, +%d Pontos de Runas)", len(resp.Feitos), resp.Xp, resp.Pontos))
		b.mu.Lock()
		if resp.State != nil {
			b.state = resp.State
		}
		b.mu.Unlock()
		if resp.State != nil {
			b.refreshStatusFromState(resp.State)
		}
	}
}

// manageRunas adquire e equipa runas passivas de XP e Loot usando os pontos de PokéLog
func (b *Bot) manageRunas(ctx context.Context) {
	runasInfo, err := b.api.GetRunas(ctx)
	if err != nil || runasInfo == nil {
		return
	}

	pontos := runasInfo.Pontos
	runaDefs := []struct {
		id    string
		custo int
	}{
		{"powerful", 550},
		{"sacred", 650},
		{"heroic", 750},
		{"victory", 850},
	}

	compradas := make(map[string]bool)
	for _, r := range runasInfo.Compradas {
		compradas[r] = true
	}
	for _, s := range runasInfo.Slots {
		if s.Runa != "" {
			compradas[s.Runa] = true
		}
	}

	livreSlot := -1
	for i, s := range runasInfo.Slots {
		if s.Runa == "" {
			livreSlot = i
			break
		}
	}

	// Se tiver runa comprada e slot livre, equipa
	if livreSlot >= 0 && len(runasInfo.Compradas) > 0 {
		r := runasInfo.Compradas[0]
		elem := "normal"
		b.mu.RLock()
		if b.currentHunt.Slug != "" && len(b.currentHunt.Pokes) > 0 {
			if pe, ok := GetPokemonEntry(b.currentHunt.Pokes[0]); ok && len(pe.Tipos) > 0 {
				elem = pe.Tipos[0]
			}
		}
		b.mu.RUnlock()

		_ = b.api.EncaixarRuna(ctx, livreSlot, r, elem)
		b.addLog(fmt.Sprintf("✨ Runa '%s' ativada no Slot %d contra elemento %s!", r, livreSlot+1, elem))
		return
	}

	// Se tiver pontos suficientes para comprar uma nova runa
	for _, rd := range runaDefs {
		if !compradas[rd.id] && pontos >= rd.custo {
			b.addLog(fmt.Sprintf("🛒 Comprando Runa '%s' por %d Pontos de PokéLog...", rd.id, rd.custo))
			if err := b.api.ComprarRuna(ctx, rd.id); err == nil {
				if livreSlot >= 0 {
					elem := "normal"
					_ = b.api.EncaixarRuna(ctx, livreSlot, rd.id, elem)
					b.addLog(fmt.Sprintf("✨ Runa '%s' ativada no Slot %d!", rd.id, livreSlot+1))
				}
				return
			}
		}
	}
}

// manageEvolucao evolui Pokémon que atingiram o nível e possuem pedras necessárias
func (b *Bot) manageEvolucao(ctx context.Context) {
	b.mu.RLock()
	st := b.state
	b.mu.RUnlock()
	if st == nil {
		return
	}

	for _, p := range st.Pokemon {
		evo, hasEvo := GetEvolucao(p.Especie)
		if !hasEvo || len(evo.Opcoes) == 0 {
			continue
		}

		if p.Nivel < evo.Nivel {
			continue
		}

		primeiraOpcao := evo.Opcoes[0]
		if nextEntry, ok := GetPokemonEntry(primeiraOpcao.Para); ok {
			if nextEntry.NivelReq > st.Nivel {
				continue
			}
		}

		podeEvoluir := true
		for pedra, qtd := range primeiraOpcao.Pedras {
			if st.Itens[pedra] < qtd {
				podeEvoluir = false
				break
			}
		}

		if podeEvoluir {
			b.addLog(fmt.Sprintf("🧬 Evoluindo Pokémon #%d (%s -> %s)...", p.ID, p.Especie, primeiraOpcao.Para))
			resp, err := b.api.EvolvePokemon(ctx, p.ID, 0)
			if err == nil {
				b.addLog(fmt.Sprintf("🎉 EVOLUÇÃO CONCLUÍDA: %s agora é %s!", p.Especie, primeiraOpcao.Para))
				if resp != nil && resp.State != nil {
					b.mu.Lock()
					b.state = resp.State
					b.mu.Unlock()
					b.refreshStatusFromState(resp.State)
				}
				return
			}
		}
	}
}

// manageGinasios desafia ginásios para conquistar insígnias e bônus passivos de tipo
func (b *Bot) manageGinasios(ctx context.Context) {
	b.mu.RLock()
	st := b.state
	b.mu.RUnlock()
	if st == nil {
		return
	}

	ginResp, err := b.api.GetGinasios(ctx)
	if err != nil || ginResp == nil {
		return
	}

	for _, g := range ginResp.Lista {
		if g.TemInsignia {
			continue
		}

		if ginResp.MeuNivel >= g.NivelMin && g.Tentativas.GratisRestam > 0 {
			b.addLog(fmt.Sprintf("🏛️ Desafiando Ginásio %s (%s)...", g.Nome, g.Tipo))
			resp, err := b.api.DesafiarGinasio(ctx, g.Tipo)
			if err == nil && resp != nil {
				if resp.Venceu {
					b.addLog(fmt.Sprintf("🏆 VITÓRIA NO GINÁSIO: Insígnia do tipo %s conquistada!", g.Tipo))
					if resp.State != nil {
						b.mu.Lock()
						b.state = resp.State
						b.mu.Unlock()
						b.refreshStatusFromState(resp.State)
					}
				} else {
					b.addLog(fmt.Sprintf("⚔️ Batalha no Ginásio %s finalizada.", g.Nome))
				}
				return
			}
		}
	}
}

// manageBoss participa dos desafios diários de World Boss
func (b *Bot) manageBoss(ctx context.Context) {
	b.mu.RLock()
	st := b.state
	b.mu.RUnlock()
	if st == nil {
		return
	}

	chefe := "magcargo"
	info, err := b.api.GetBossInfo(ctx, chefe)
	if err != nil || info == nil {
		return
	}

	if info.Entradas.GratisRestam > 0 {
		activePoke := b.getActivePokemon()
		pokeID := 0
		if activePoke != nil {
			pokeID = activePoke.ID
		}
		b.addLog(fmt.Sprintf("👹 Participando do World Boss: %s (Entrada Grátis)...", chefe))
		resp, err := b.api.LutarBoss(ctx, chefe, pokeID)
		if err == nil && resp != nil {
			if resp.Venceu {
				b.addLog(fmt.Sprintf("🎉 CHEFE DERROTADO: +%d XP, +%d Ouro!", resp.Xp, resp.Gold))
			} else {
				b.addLog("⚔️ Luta contra World Boss finalizada.")
			}
		}
	}
}

// manageAutoVIP executa a estratégia de compra automática de VIP:
// Fase 1: Adquire a 1ª VIP de 30 dias (25 diamantes). Se faltar diamantes, compra no mercado com ouro.
// Fase 2: Com a 1ª VIP ativa, acumula e adquire o plano semestral com desconto (180 dias / 135 diamantes).
func (b *Bot) manageAutoVIP(ctx context.Context) {
	b.mu.RLock()
	st := b.state
	b.mu.RUnlock()
	if st == nil {
		return
	}

	me, err := b.api.GetMe(ctx)
	if err != nil || me == nil {
		return
	}

	totalDiamonds := me.Diamonds + me.DiamondsBound

	if !me.Vip {
		// Fase 1: Alvo = 25 diamantes para 30 dias de VIP
		alvoDiamantes := 25
		if totalDiamonds >= alvoDiamantes {
			b.addLog("👑 Saldo suficiente! Ativando VIP de 30 dias...")
			resp, err := b.api.BuyVIP(ctx, 30)
			if err == nil && resp != nil {
				b.addLog("🎉 VIP DE 30 DIAS ATIVADA COM SUCESSO! (+50% EXP Treinador/Pokémon, +15% Loot, Filtros Auto-Catch)!")
				b.mu.Lock()
				if b.state != nil {
					b.state.Vip = true
				}
				if b.status.Extra == nil {
					b.status.Extra = make(map[string]any)
				}
				b.status.Extra["vip"] = true
				b.mu.Unlock()
			}
			return
		}

		// Se faltar diamantes e tivermos ouro suficiente, busca ofertas no mercado
		falta := alvoDiamantes - totalDiamonds
		mkt, err := b.api.GetMarketListings(ctx, "diamante")
		if err == nil && mkt != nil && len(mkt.Anuncios) > 0 {
			var bestListing *MarketListingItem
			var bestUnitCost int64 = 0

			for i := range mkt.Anuncios {
				item := &mkt.Anuncios[i]
				if item.Tipo != "diamante" || item.Moeda != "dolar" || item.Preco <= 0 {
					continue
				}
				unit := item.Preco
				if item.ItemQtd > 0 {
					unit = item.Preco / int64(item.ItemQtd)
				}
				if bestListing == nil || unit < bestUnitCost {
					bestListing = item
					bestUnitCost = unit
				}
			}

			if bestListing != nil && bestUnitCost > 0 {
				custoTotal := bestUnitCost * int64(falta)
				minGold := b.config.Settings.MinGoldKeep
				if minGold <= 0 {
					minGold = 50000
				}
				if st.Gold >= (custoTotal + minGold) {
					b.addLog(fmt.Sprintf("🛒 Comprando %d Diamantes no Mercado por %d de Ouro para ativar VIP...", falta, custoTotal))
					compraQtd := falta
					if !bestListing.Avulso {
						compraQtd = bestListing.ItemQtd
					}
					buyResp, err := b.api.BuyMarket(ctx, bestListing.ID, compraQtd)
					if err == nil && buyResp != nil {
						b.addLog("💎 Compra de Diamantes concluída! Ativando VIP de 30 dias...")
						vipResp, err := b.api.BuyVIP(ctx, 30)
						if err == nil && vipResp != nil {
							b.addLog("🎉 VIP DE 30 DIAS ATIVADA COM SUCESSO! (+50% EXP, +15% Loot, Auto-Catch avançado)!")
							b.mu.Lock()
							if b.state != nil {
								b.state.Vip = true
							}
							if b.status.Extra == nil {
								b.status.Extra = make(map[string]any)
							}
							b.status.Extra["vip"] = true
							b.mu.Unlock()
						}
					}
				}
			}
		}
	} else {
		// Fase 2: Já é VIP! Acumular para o plano com desconto máximo (180 dias por 135 diamantes)
		if totalDiamonds >= 135 {
			b.addLog("👑 Saldo suficiente para pacote semestral! Renovando VIP por 180 dias...")
			resp, err := b.api.BuyVIP(ctx, 180)
			if err == nil && resp != nil {
				b.addLog("🎉 VIP DE 180 DIAS (6 MESES) ATIVADA! Estabilidade máxima garantida!")
			}
		} else if totalDiamonds >= 48 {
			b.addLog("👑 Renovando VIP por 60 dias (2 meses)...")
			resp, err := b.api.BuyVIP(ctx, 60)
			if err == nil && resp != nil {
				b.addLog("🎉 VIP RENOVADA POR 60 DIAS COM SUCESSO!")
			}
		}
	}
}

func findActivePokemon(st *PokeTrueState) *PokemonState {
	if st == nil {
		return nil
	}
	for i := range st.Pokemon {
		if st.Pokemon[i].ID == st.Hunt.AtivoID {
			return &st.Pokemon[i]
		}
	}
	if len(st.Pokemon) > 0 {
		return &st.Pokemon[0]
	}
	return nil
}

func (b *Bot) getActivePokemon() *PokemonState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return findActivePokemon(b.state)
}

func (b *Bot) getActivePokemonHP() int {
	poke := b.getActivePokemon()
	if poke != nil {
		return int(poke.Hp)
	}
	return 100
}

func (b *Bot) refreshStatusFromState(st *PokeTrueState) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.status.Level = st.Nivel
	b.status.Gold = st.Gold
	b.status.CurrentArea = st.Hunt.Mapa
	b.status.UpdatedAt = time.Now().UnixMilli()

	active := findActivePokemon(st)
	if active != nil {
		b.status.Vocation = active.Especie
		b.status.HP = int(active.Hp)
		b.status.MaxHP = active.HpMax
		if b.status.MaxHP <= 0 {
			b.status.MaxHP = 100
		}
	}

	// Calcular EXP/h dinâmico
	elapsedHours := time.Since(b.startTime).Hours()
	if elapsedHours > 0.005 {
		xpGained := st.NivelXp - b.initialXp
		if xpGained > 0 {
			b.status.ExpPerHour = float64(xpGained) / elapsedHours
		}
	}
	b.status.DurationSec = int(time.Since(b.startTime).Seconds())
}

func (b *Bot) addLog(msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	log.Printf("[%s] %s", b.config.ID, msg)
	b.status.Logs = append(b.status.Logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	if len(b.status.Logs) > 15 {
		b.status.Logs = b.status.Logs[1:]
	}
}

func (b *Bot) updateStatus(fn func(*engine.AccountStatus)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	fn(&b.status)
	b.status.UpdatedAt = time.Now().UnixMilli()
}

func (b *Bot) GetStatus() engine.AccountStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

func (b *Bot) ExecuteAction(action string, payload map[string]any) (any, error) {
	switch action {
	case "setHunt":
		if target, ok := payload["target"].(string); ok {
			b.mu.Lock()
			b.config.Settings.HuntTarget = target
			b.mu.Unlock()
			b.addLog(fmt.Sprintf("Área alterada manualmente para: %s", target))
			return "Área alterada com sucesso", nil
		}
		return nil, fmt.Errorf("parâmetro 'target' ausente")

	case "autoHunt":
		b.mu.Lock()
		b.config.Settings.AutoHunt = !b.config.Settings.AutoHunt
		val := b.config.Settings.AutoHunt
		b.mu.Unlock()
		return map[string]any{"autoHunt": val}, nil

	case "autoSell":
		go b.manageAutoSell(context.Background())
		return "Auto-sell disparado", nil

	case "autoEquip":
		go b.manageHelds(context.Background())
		return "Verificação de helds disparada", nil

	case "autoHeal":
		go b.manageAutoHeal(context.Background())
		return "Auto-heal executado", nil

	case "sync":
		st, err := b.api.GetState(context.Background())
		if err != nil {
			return nil, err
		}
		b.mu.Lock()
		b.state = st
		b.mu.Unlock()
		b.refreshStatusFromState(st)
		return "Estado sincronizado com o servidor", nil

	default:
		return nil, fmt.Errorf("ação desconhecida: %s", action)
	}
}

func (b *Bot) SupplyTicket(ticket, wsURL string) error {
	b.api.SetToken(ticket)
	b.addLog("Token entregue via SupplyTicket.")
	return nil
}

func (b *Bot) SupplyTurnstileToken(token string) error {
	select {
	case b.turnstileChan <- token:
	default:
	}
	return nil
}
