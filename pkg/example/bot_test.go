package example

import (
	"context"
	"testing"
	"time"
)

func TestBotLifecycle(t *testing.T) {
	cfg := AccountConfig{
		ID:      "test-acc",
		Name:    "TestHero",
		Enabled: true,
		Settings: GameSettings{
			HuntTarget: "auto",
			AutoHunt:   true,
			AutoDaily:  true,
		},
	}

	bot := NewBot(cfg)
	if bot.ID() != "test-acc" {
		t.Errorf("ID incorreto: %s", bot.ID())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := bot.Init(ctx); err != nil {
		t.Fatalf("Init falhou: %v", err)
	}

	_ = bot.Run(ctx)

	st := bot.GetStatus()
	if st.ID != "test-acc" {
		t.Errorf("Status ID inesperado: %s", st.ID)
	}
}
