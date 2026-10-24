package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"bot-template-go/pkg/engine"
	"bot-template-go/pkg/game"
	"bot-template-go/pkg/logger"
	"bot-template-go/pkg/server"
)

//go:embed accounts.example.json
var defaultAccountsJSON []byte

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1_000_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%.2fkk", float64(n)/1_000_000.0)
}

func main() {
	logger.InitLogger(slog.LevelInfo)

	fmt.Println("================================================================================")
	fmt.Println("       ⚡ NATIVE GO MULTI-BOT ENGINE (ULTRA-LEVE & ALTA PERFORMANCE) ⚡        ")
	fmt.Println("================================================================================")

	// 1. Carregar contas do accounts.json (cria a partir do template se não existir)
	configFile := "accounts.json"
	content, err := os.ReadFile(configFile)
	if err != nil {
		slog.Warn("Arquivo accounts.json não encontrado. Criando template padrão...")
		_ = os.WriteFile(configFile, defaultAccountsJSON, 0644)
		content = defaultAccountsJSON
	}

	var accountConfigs []game.AccountConfig
	if err := json.Unmarshal(content, &accountConfigs); err != nil {
		slog.Error("Falha crítica ao analisar accounts.json", "erro", err)
		os.Exit(1)
	}

	targetAcc := os.Getenv("ACCOUNT")
	var activeConfigs []game.AccountConfig

	for _, a := range accountConfigs {
		if !a.Enabled && a.ID != targetAcc {
			continue
		}
		if targetAcc != "" && a.ID != targetAcc {
			continue
		}
		activeConfigs = append(activeConfigs, a)
	}

	if len(activeConfigs) == 0 {
		slog.Warn("Nenhuma conta com 'enabled: true' encontrada em accounts.json")
		slog.Info("Edite accounts.json e habilite sua conta para iniciar.")
	}
	if err := game.ValidateActiveAccounts(activeConfigs); err != nil {
		slog.Error("Configuração de contas inválida", "erro", err)
		os.Exit(1)
	}

	// Context com graceful shutdown nativo
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	manager := engine.NewBotManager()

	slog.Info("Iniciando contas em Goroutines isoladas com supervisor", "total", len(activeConfigs))
	for _, cfg := range activeConfigs {
		b := game.NewBot(cfg)
		if err := manager.Register(b); err != nil {
			slog.Error("Conta ignorada por erro de registro", "id", cfg.ID, "erro", err)
			continue
		}
		go engine.RunWithSupervisor(ctx, b)
	}

	// 2. Iniciar Servidor Web / Dashboard HTTP
	webPort := 3001
	if p := os.Getenv("PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			webPort = v
		}
	}

	srv := server.NewServer(webPort, manager)
	go func() {
		slog.Info("Dashboard online", "url", fmt.Sprintf("http://localhost:%d", webPort))
		if err := srv.Start(); err != nil && err.Error() != "http: Server closed" {
			slog.Warn("Web Server encerrado", "erro", err)
		}
	}()

	// 3. Monitor no Terminal (atualiza a cada 4 segundos com estatísticas de RAM)
	go func() {
		time.Sleep(2 * time.Second)
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				allocMB := float64(m.Alloc) / 1024.0 / 1024.0
				sysMB := float64(m.Sys) / 1024.0 / 1024.0

				statuses := manager.GetAllStatuses()
				var totalGold int64 = 0
				totalKills := 0
				huntingCount := 0

				for _, st := range statuses {
					totalGold += st.Gold
					totalKills += st.Kills
					if st.Status == "hunting" {
						huntingCount++
					}
				}

				fmt.Print("\033[H\033[2J") // Limpar tela do terminal
				fmt.Println("================================================================================")
				fmt.Printf("   ⚡ BOT ENGINE CONSOLE — RAM: %.2f MB (Sys: %.2f MB) | Goroutines: %d\n", allocMB, sysMB, runtime.NumGoroutine())
				fmt.Printf("   📊 Contas: %d/%d Caçando | Ouro: %s | Total Kills: %d\n", huntingCount, len(statuses), formatNumber(totalGold), totalKills)
				fmt.Printf("   🌐 Dashboard Web: http://localhost:%d\n", webPort)
				fmt.Println("================================================================================")

				for _, st := range statuses {
					hpPct := 0
					if st.MaxHP > 0 {
						hpPct = (st.HP * 100) / st.MaxHP
					}
					manaPct := 0
					if st.MaxMana > 0 {
						manaPct = (st.Mana * 100) / st.MaxMana
					}

					fmt.Printf("▸ [%s] %s (Lv. %d %s) | Status: %s\n", st.ID, st.Name, st.Level, st.Vocation, st.Status)
					fmt.Printf("  HP: %d/%d (%d%%) | Mana: %d/%d (%d%%) | Ouro: %s | Kills: %d\n", st.HP, st.MaxHP, hpPct, st.Mana, st.MaxMana, manaPct, formatNumber(st.Gold), st.Kills)
					if st.CurrentArea != "" {
						fmt.Printf("  Área: %s | EXP/H: %.0f | Duração: %ds\n", st.CurrentArea, st.ExpPerHour, st.DurationSec)
					}
					fmt.Println("--------------------------------------------------------------------------------")
				}
			}
		}
	}()

	// 4. Bloquear até receber sinal SIGINT/SIGTERM
	<-ctx.Done()
	slog.Info("Sinal de encerramento recebido. Desligando graciosamente...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	_ = srv.Stop(shutdownCtx)

	slog.Info("Todos os bots e serviços foram desligados com sucesso.")
}
