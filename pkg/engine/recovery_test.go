package engine

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func expireCooldown(cb *CircuitBreaker) {
	cb.mu.Lock()
	cb.openedAt = time.Now().Add(-cb.currentWait)
	cb.mu.Unlock()
}

func TestRecoveryRequiresExclusiveCanonicalProbe(t *testing.T) {
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())
	cb.Trip("rate limit")
	cb.RecordSuccess()
	if cb.State() != StateOpen {
		t.Fatal("resposta antiga encerrou o cooldown")
	}
	expireCooldown(cb)
	cb.RecordSuccess()
	if cb.AllowAction() {
		t.Fatal("progresso permitido sem recuperar estado")
	}
	var acquired atomic.Int32
	var generation atomic.Uint64
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if gen, ok := cb.BeginRecovery(); ok {
				acquired.Add(1)
				generation.Store(gen)
			}
		}()
	}
	wg.Wait()
	if acquired.Load() != 1 {
		t.Fatalf("esperava uma leitura, obteve %d", acquired.Load())
	}
	cb.FinishRecovery(generation.Load()+1, nil)
	if cb.State() != StateHalfOpen {
		t.Fatal("resultado de outra geração foi aceito")
	}
	cb.RecordSuccess()
	if cb.State() != StateHalfOpen {
		t.Fatal("resposta comum validou a recuperação")
	}
	cb.FinishRecovery(generation.Load(), errors.New("JSON inválido"))
	if cb.State() != StateOpen || cb.currentWait != 30*time.Minute {
		t.Fatal("falha de recuperação não aplicou novo cooldown com backoff")
	}
	cb.FinishRecovery(generation.Load(), nil)
	if cb.State() != StateOpen {
		t.Fatal("resultado antigo reabriu o circuito")
	}
}

func TestTripCallbackIsImmediateAndCanInspectBreaker(t *testing.T) {
	var cb *CircuitBreaker
	called := 0
	cfg := DefaultCircuitBreakerConfig()
	cfg.OnTrip = func(string) {
		called++
		if cb.State() != StateOpen {
			t.Error("callback deve observar o circuito aberto")
		}
	}
	cb = NewCircuitBreaker("test", cfg)
	cb.Trip("primeira recusa")
	openedAt := cb.openedAt
	cb.Trip("resposta concorrente")
	if called != 1 || cb.openedAt != openedAt {
		t.Fatal("Trip repetido deve preservar o cooldown e executar o callback uma vez")
	}
}

func TestRejectionJSONFormatting(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		body     string
		rejected bool
	}{
		{"403 sem Cloudflare", 403, "Forbidden", true},
		{"429 vazio", 429, "", true},
		{"flag com quebra de linha", 200, "{\"recusado\" :\n\ttrue}", true},
		{"flag aninhada", 200, `{"result":{"outraAba":true}}`, true},
		{"flag escapada", 200, `{"\u0062anido":true}`, true},
		{"mensagem escapada", 200, `{"erro":"experi\u00eancia r\u00e1pida demais"}`, true},
		{"flags falsas", 200, `{"banido":false,"outraAba":false,"recusado":false}`, false},
		{"string não é flag", 200, `{"recusado":"true"}`, false},
		{"sucesso", 200, `{"ok":true}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, reason := IsAntiCheatRejection(tc.status, tc.body)
			if got != tc.rejected {
				t.Fatalf("rejeição = %v, motivo = %s", got, reason)
			}
		})
	}
}
