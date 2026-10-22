package game

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"bot-template-go/pkg/engine"
)

func localClient(t *testing.T, cb *engine.CircuitBreaker, handler http.HandlerFunc) *PokeTrueClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewPokeTrueClient("test-token", "", cb)
	c.baseURL = srv.URL + "/"
	t.Cleanup(c.httpClient.CloseIdleConnections)
	return c
}

func waitHalfOpen(t *testing.T, cb *engine.CircuitBreaker) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for cb.State() == engine.StateOpen {
		if time.Now().After(deadline) {
			t.Fatal("cooldown não expirou")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestClientBlocksQueuedProgressAfterRejection(t *testing.T) {
	cb := engine.NewCircuitBreaker("test", engine.DefaultCircuitBreakerConfig())
	var progress, stops atomic.Int32
	c := localClient(t, cb, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hunt/stop" {
			stops.Add(1)
			fmt.Fprint(w, `{}`)
			return
		}
		progress.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	})
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.StopGrupo(context.Background())
			if !errors.Is(err, ErrAntiCheat) && !errors.Is(err, ErrCircuitOpen) {
				t.Errorf("erro inesperado: %v", err)
			}
		}()
	}
	wg.Wait()
	if progress.Load() != 1 || stops.Load() != 1 || cb.State() != engine.StateOpen {
		t.Fatalf("progress=%d stop=%d state=%v", progress.Load(), stops.Load(), cb.State())
	}
}

func TestClientRecoveryOnlyAcceptsValidCanonicalState(t *testing.T) {
	for _, body := range []string{`{"nivel":1,"gold":10}`, `{}`, `null`, `{"erro":"sessao_expirada"}`, `invalid`, ""} {
		t.Run(body, func(t *testing.T) {
			cfg := engine.DefaultCircuitBreakerConfig()
			cfg.BaseCooldown = 50 * time.Millisecond
			cb := engine.NewCircuitBreaker("test", cfg)
			var calls atomic.Int32
			c := localClient(t, cb, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != http.MethodGet || r.URL.Path != "/state" {
					t.Error("recuperação usou ação de progresso")
				}
				fmt.Fprint(w, body)
			})
			cb.Trip("test")
			waitHalfOpen(t, cb)
			if err := c.StopGrupo(context.Background()); !errors.Is(err, ErrCircuitOpen) {
				t.Fatal("progresso permitido em half-open")
			}
			_, err := c.GetState(context.Background())
			valid := body == `{"nivel":1,"gold":10}`
			if valid != (err == nil) {
				t.Fatalf("estado válido=%v erro=%v", valid, err)
			}
			if valid && cb.State() != engine.StateClosed {
				t.Fatal("recuperação válida não fechou circuito")
			}
			if !valid && cb.State() != engine.StateOpen {
				t.Fatal("recuperação inválida não reabriu circuito")
			}
			if calls.Load() != 1 {
				t.Fatal("requisições extras durante recuperação")
			}
		})
	}
}

func TestClientInstanceAndTokenIsolation(t *testing.T) {
	var instance string
	var instanceMu sync.Mutex
	c := localClient(t, nil, func(w http.ResponseWriter, r *http.Request) {
		instanceMu.Lock()
		defer instanceMu.Unlock()
		id := r.Header.Get("X-Instancia")
		if id == "" || (instance != "" && id != instance) {
			t.Error("instância ausente ou instável")
		}
		instance = id
		if r.Header.Get("Authorization") != "Bearer refreshed" {
			t.Error("token não atualizado")
		}
		fmt.Fprint(w, `{}`)
	})
	c.SetToken("refreshed")
	for range 2 {
		if err := c.StopHunt(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	other := NewPokeTrueClient("", "", nil)
	instanceMu.Lock()
	defer instanceMu.Unlock()
	if other.windowID == instance {
		t.Fatal("clientes compartilham instância")
	}
}

func TestClientDoesNotTreatLogicalErrorsAsSuccess(t *testing.T) {
	c := localClient(t, nil, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"erro":"saldo_insuficiente"}`) })
	if err := c.BuyItem(context.Background(), "potion", 1); err == nil {
		t.Fatal("erro lógico foi ignorado")
	}
}

func TestBotTripDisablesHuntingAndSnapshotsAreIndependent(t *testing.T) {
	b := NewBot(AccountConfig{ID: "test", Settings: GameSettings{AutoHunt: true}})
	b.cb.Trip("test")
	if b.settings().AutoHunt || b.GetStatus().Status != "cooldown" {
		t.Fatal("callback de proteção não foi instalado")
	}
	if _, err := b.ExecuteAction("autoHunt", nil); !errors.Is(err, ErrCircuitOpen) {
		t.Fatal("caça reativada durante cooldown")
	}
	snapshot := b.GetStatus()
	snapshot.Logs[0] = "changed"
	snapshot.Extra["isResting"] = false
	if b.GetStatus().Logs[0] == "changed" || b.GetStatus().Extra["isResting"] != true {
		t.Fatal("GetStatus compartilha memória mutável")
	}
}

func TestAutoHuntDisabledDoesNotSendProgress(t *testing.T) {
	b := NewBot(AccountConfig{ID: "test"})
	b.state = &PokeTrueState{Nivel: 1}
	b.api = localClient(t, b.cb, func(w http.ResponseWriter, r *http.Request) { t.Error("caça desabilitada enviou requisição") })
	if err := b.performBotStep(context.Background()); err != nil {
		t.Fatal(err)
	}
}
