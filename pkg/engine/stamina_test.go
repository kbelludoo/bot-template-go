package engine

import (
	"testing"
	"time"
)

func TestStaminaControllerFatigue(t *testing.T) {
	cfg := DefaultStaminaConfig(StaminaTypeIncreasingFatigue)
	cfg.MinRestDuration = 50 * time.Millisecond // Para teste rápido

	sc := NewStaminaController(cfg)

	// Estado inicial: 10% fadiga, 60min farm
	sc.UpdateFatigue(100, 3600000, 35.0)

	should, _ := sc.ShouldRest()
	if should {
		t.Errorf("Não deveria descansar com 10%% de fadiga")
	}

	// Sobe fadiga para 85%
	sc.UpdateFatigue(850, 1800000, 20.0)
	should, reason := sc.ShouldRest()
	if !should {
		t.Errorf("Deveria descansar com 85%% de fadiga")
	}
	t.Logf("Descanso acionado com sucesso: %s", reason)

	sc.EnterRest()
	if !sc.IsResting() {
		t.Errorf("Deveria estar em modo descanso")
	}

	// Tentar retomar imediatamente (deve ser bloqueado pela histerese)
	can, _ := sc.CanResume()
	if can {
		t.Errorf("Não deveria poder retomar antes do tempo mínimo de descanso")
	}

	// Aguardar histerese
	time.Sleep(60 * time.Millisecond)

	// Ainda com 85% de fadiga, não pode retomar
	can, _ = sc.CanResume()
	if can {
		t.Errorf("Não deveria retomar ainda com 85%% de fadiga")
	}

	// Fadiga cai para 25% e farm disponível sobe para 15 min
	sc.UpdateFatigue(250, 900000, 30.0)
	can, resumeReason := sc.CanResume()
	if !can {
		t.Errorf("Deveria poder retomar após fadiga cair para 25%%")
	}
	t.Logf("Retomada autorizada com sucesso: %s", resumeReason)

	sc.ExitRest()
	if sc.IsResting() {
		t.Errorf("Não deveria estar descansando após ExitRest")
	}

	extra := sc.ExportExtra()
	if extra["fatiguePercent"] != 25.0 {
		t.Errorf("Esperado fatiguePercent 25.0, obtido %v", extra["fatiguePercent"])
	}
}

func TestStaminaControllerDecreasing(t *testing.T) {
	cfg := DefaultStaminaConfig(StaminaTypeDecreasing)
	cfg.MinRestDuration = 10 * time.Millisecond

	sc := NewStaminaController(cfg)
	sc.UpdateStaminaHours(42.0)

	should, _ := sc.ShouldRest()
	if should {
		t.Errorf("Não deveria descansar com 42h")
	}

	sc.UpdateStaminaHours(14.0)
	should, reason := sc.ShouldRest()
	if !should {
		t.Errorf("Deveria descansar com 14h")
	}
	t.Logf("Descanso por stamina baixa: %s", reason)

	sc.EnterRest()
	time.Sleep(20 * time.Millisecond)

	sc.UpdateStaminaHours(39.0)
	can, reason := sc.CanResume()
	if !can {
		t.Errorf("Deveria poder retomar com 39h")
	}
	t.Logf("Retomada por stamina verde: %s", reason)
}
