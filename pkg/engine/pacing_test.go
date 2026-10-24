package engine

import (
	"testing"
	"time"
)

func TestAdaptiveSpeedTuner(t *testing.T) {
	cfg := DefaultAdaptiveSpeedConfig()
	cfg.InitialSpeed = 1.20
	cfg.CleanRunsRequired = 2
	cfg.StepUp = 0.05
	cfg.StepDown = 0.10

	tuner := NewAdaptiveSpeedTuner(cfg)
	if tuner.Speed() != 1.20 {
		t.Fatalf("esperava velocidade inicial 1.20, obteve %v", tuner.Speed())
	}

	// 2 clean runs -> sobe +0.05
	tuner.RecordCleanRun()
	s := tuner.RecordCleanRun()
	if s != 1.25 {
		t.Fatalf("esperava 1.25 após 2 runs limpas, obteve %v", s)
	}

	// 1 throttle -> reduz -0.10
	s = tuner.RecordThrottle()
	if s != 1.15 {
		t.Fatalf("esperava 1.15 após throttle, obteve %v", s)
	}

	// Duração de run: meta 100, taxa 5/s, speed 1.15 -> 100 / 5.75 = 17.39s
	dur := tuner.CalculateRunDuration(100, 5.0, 8.0)
	if dur < 17*time.Second || dur > 18*time.Second {
		t.Fatalf("duração inesperada: %v", dur)
	}
}

func TestAntiWipeTracker(t *testing.T) {
	tracker := NewAntiWipeTracker(3, 10*time.Minute, 10*time.Minute)
	base := time.Now()

	if tracker.RecordDeath(base) {
		t.Fatal("não deveria rebaixar na 1ª morte")
	}
	if tracker.RecordDeath(base.Add(2 * time.Minute)) {
		t.Fatal("não deveria rebaixar na 2ª morte")
	}
	if !tracker.RecordDeath(base.Add(4 * time.Minute)) {
		t.Fatal("deveria rebaixar na 3ª morte dentro de 10 minutos")
	}
	if !tracker.IsDowngraded(base.Add(5 * time.Minute)) {
		t.Fatal("deveria continuar rebaixado aos 5m")
	}
	if tracker.IsDowngraded(base.Add(15 * time.Minute)) {
		t.Fatal("deveria restaurar após 10m de cooldown sem mortes")
	}
}

func TestRankingStealthController(t *testing.T) {
	ctrl := NewRankingStealthController(5, 30, 0.5)
	ctrl.UpdateTopRankingLevel(100)

	if freeze, _ := ctrl.ShouldFreezeXP(94); freeze {
		t.Fatal("nível 94 não deve congelar XP quando top é 100 (margem 5 -> teto 95)")
	}
	if freeze, capLvl := ctrl.ShouldFreezeXP(95); !freeze || capLvl != 95 {
		t.Fatalf("nível 95 deveria congelar XP no teto 95 (got freeze=%v cap=%d)", freeze, capLvl)
	}
	if mult := ctrl.KillRateMultiplier(25); mult != 1.0 {
		t.Fatalf("abaixo do nível 30 mult deve ser 1.0, obteve %v", mult)
	}
	if mult := ctrl.KillRateMultiplier(30); mult != 0.5 {
		t.Fatalf("no nível 30+ mult deve ser 0.5, obteve %v", mult)
	}
}

func TestCircularPatrolPosition(t *testing.T) {
	now := time.UnixMilli(1700000000000)
	x1, y1 := CircularPatrolPosition(50, 50, 3.0, 0.6, "Conta1", now)
	x2, y2 := CircularPatrolPosition(50, 50, 3.0, 0.6, "Conta2", now)

	if x1 == x2 && y1 == y2 {
		t.Fatalf("dois bots com nicks distintos não devem empilhar no mesmo pixel: (%v,%v)", x1, y1)
	}
}
