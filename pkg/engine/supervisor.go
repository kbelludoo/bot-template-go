package engine

import (
	"context"
	"fmt"
	"log"
	"time"
)

// RunWithSupervisor roda um bot continuamente, reiniciando com backoff se houver queda ou panic
func RunWithSupervisor(ctx context.Context, b GameBot) {
	backoff := 3 * time.Second
	const maxBackoff = 30 * time.Second

	for ctx.Err() == nil {
		started := time.Now()
		err := runBotSession(ctx, b)
		if ctx.Err() != nil {
			return
		}
		// Uma conexão curta não deve zerar o backoff e gerar um loop de reconexão.
		if time.Since(started) >= time.Minute {
			backoff = 3 * time.Second
		}
		log.Printf("[%s] Sessão encerrada: %v. Reiniciando em %v...", b.ID(), err, backoff)
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func runBotSession(ctx context.Context, b GameBot) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic recuperado: %v", recovered)
		}
	}()
	log.Printf("[%s] Inicializando bot...", b.ID())
	if err := b.Init(ctx); err != nil {
		return fmt.Errorf("Init: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return b.Run(ctx)
}
