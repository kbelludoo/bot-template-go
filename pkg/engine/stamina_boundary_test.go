package engine

import (
	"math"
	"testing"
	"time"
)

func TestStaminaMissingAndExhaustedMetrics(t *testing.T) {
	for _, kind := range []StaminaType{StaminaTypeIncreasingFatigue, StaminaTypeDecreasing} {
		s := NewStaminaController(DefaultStaminaConfig(kind))
		if rest, _ := s.ShouldRest(); !rest {
			t.Fatal("não deve caçar antes da primeira métrica")
		}
	}
	s := NewStaminaController(DefaultStaminaConfig(StaminaTypeIncreasingFatigue))
	for _, farm := range []uint64{0, 119999} {
		s.UpdateFatigue(0, farm, 0)
		if rest, _ := s.ShouldRest(); !rest {
			t.Fatalf("farm disponível %d deve exigir descanso", farm)
		}
	}
	s.UpdateFatigue(0, 120000, 0)
	if rest, _ := s.ShouldRest(); rest {
		t.Fatal("120000ms ainda permite operação")
	}
	s = NewStaminaController(DefaultStaminaConfig(StaminaTypeDecreasing))
	for _, hours := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		s.UpdateStaminaHours(hours)
		if rest, _ := s.ShouldRest(); !rest {
			t.Fatalf("stamina %v deve impedir caça", hours)
		}
	}
}

func TestFatigueResumeRequiresBothStrictThresholds(t *testing.T) {
	s := NewStaminaController(DefaultStaminaConfig(StaminaTypeIncreasingFatigue))
	s.EnterRest()
	s.restStartTime = time.Now().Add(-10 * time.Minute)
	started := s.restStartTime
	s.EnterRest()
	if s.restStartTime != started {
		t.Fatal("EnterRest repetido reiniciou o descanso mínimo")
	}
	for _, tc := range []struct {
		fatigue uint32
		farm    uint64
		resume  bool
	}{
		{300, 900000, false}, {299, 0, false}, {299, 600000, false}, {299, 600001, true},
	} {
		s.UpdateFatigue(tc.fatigue, tc.farm, 0)
		if resume, _ := s.CanResume(); resume != tc.resume {
			t.Fatalf("fatigue=%d farm=%d: resume=%v", tc.fatigue, tc.farm, resume)
		}
	}
}
