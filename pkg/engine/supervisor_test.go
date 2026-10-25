package engine

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubBot struct {
	id      string
	init    func(context.Context) error
	run     func(context.Context) error
	tickets int
}

func (b *stubBot) ID() string                                        { return b.id }
func (b *stubBot) Name() string                                      { return b.id }
func (b *stubBot) Init(ctx context.Context) error                    { return b.init(ctx) }
func (b *stubBot) Run(ctx context.Context) error                     { return b.run(ctx) }
func (b *stubBot) GetStatus() AccountStatus                          { return NewStatus(b.id, b.id) }
func (b *stubBot) ExecuteAction(string, map[string]any) (any, error) { return nil, nil }
func (b *stubBot) SupplyTicket(string, string) error                 { b.tickets++; return nil }
func (b *stubBot) SupplyTurnstileToken(string) error                 { b.tickets++; return nil }

func TestSupervisorCancellationDuringBackoff(t *testing.T) {
	for _, outcome := range []string{"init error", "init panic", "run error", "run panic", "run nil"} {
		t.Run(outcome, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			called := make(chan struct{})
			b := &stubBot{id: "test"}
			b.init = func(context.Context) error {
				if outcome == "init error" || outcome == "init panic" {
					close(called)
					if outcome == "init panic" {
						panic("init")
					}
					return errors.New("init")
				}
				return nil
			}
			b.run = func(context.Context) error {
				close(called)
				if outcome == "run panic" {
					panic("run")
				}
				if outcome == "run error" {
					return errors.New("run")
				}
				return nil
			}
			done := make(chan struct{})
			go func() { RunWithSupervisor(ctx, b); close(done) }()
			select {
			case <-called:
			case <-time.After(time.Second):
				t.Fatal("sessão não iniciou")
			}
			cancel()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("cancelamento bloqueado no backoff")
			}
		})
	}
}

func TestManagerRejectsDuplicateAndAmbiguousAccounts(t *testing.T) {
	m := NewBotManager()
	if statuses := m.GetAllStatuses(); statuses == nil {
		t.Fatal("status vazio deve ser []")
	}
	first, second := &stubBot{id: "first"}, &stubBot{id: "second"}
	if err := m.Register(first); err != nil {
		t.Fatal(err)
	}
	if err := m.Register(first); err == nil {
		t.Fatal("ID duplicado aceito")
	}
	if err := m.Register(&stubBot{}); err == nil {
		t.Fatal("ID vazio aceito")
	}
	if err := m.SupplyTicket("missing", "token", ""); err == nil {
		t.Fatal("conta desconhecida aceitou ticket")
	}
	if err := m.SupplyTicket("", "token", ""); err != nil {
		t.Fatal(err)
	}
	if err := m.Register(second); err != nil {
		t.Fatal(err)
	}
	if err := m.SupplyTicket("", "token", ""); err == nil {
		t.Fatal("ticket ambíguo aceito")
	}
	if err := m.SupplyTurnstileToken("missing", "token"); err == nil {
		t.Fatal("token foi entregue à conta errada")
	}
	if err := m.SupplyTurnstileToken("", "token"); err == nil {
		t.Fatal("token ambíguo aceito")
	}
	if first.tickets != 1 || second.tickets != 0 || m.Count() != 2 {
		t.Fatal("roteamento ou registro incorreto")
	}
}
