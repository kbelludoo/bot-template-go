package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestConsoleHandler(t *testing.T) {
	buf := new(bytes.Buffer)
	handler := NewConsoleHandler(buf, slog.LevelDebug)
	l := slog.New(handler)

	l.Info("Servidor iniciado", "porta", 3000)
	out := buf.String()
	if !strings.Contains(out, "Servidor iniciado") || !strings.Contains(out, "porta=3000") {
		t.Fatalf("Log inesperado: %s", out)
	}

	buf.Reset()
	botLog := l.With("bot", "acc-1")
	botLog.Warn("Stamina baixa", "horas", 13.5)
	out2 := buf.String()
	if !strings.Contains(out2, "[acc-1]") || !strings.Contains(out2, "Stamina baixa") {
		t.Fatalf("Log com bot inesperado: %s", out2)
	}
}
