package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	defaultLogger *slog.Logger
	loggerOnce    sync.Once
)

// ConsoleHandler formata logs coloridos e limpos para o terminal da VPS/Console
type ConsoleHandler struct {
	w     io.Writer
	level slog.Level
	mu    sync.Mutex
	attrs []slog.Attr
	group string
}

func NewConsoleHandler(w io.Writer, level slog.Level) *ConsoleHandler {
	return &ConsoleHandler{
		w:     w,
		level: level,
	}
}

func (h *ConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	timeStr := r.Time.Format("15:04:05")

	var levelTag string
	switch r.Level {
	case slog.LevelDebug:
		levelTag = "\033[36m[DEBUG]\033[0m"
	case slog.LevelInfo:
		levelTag = "\033[32m[INFO]\033[0m"
	case slog.LevelWarn:
		levelTag = "\033[33m[WARN]\033[0m"
	case slog.LevelError:
		levelTag = "\033[31m[ERROR]\033[0m"
	default:
		levelTag = fmt.Sprintf("[%s]", r.Level)
	}

	// Extrair atributo "bot" ou "account" se existir
	var accountTag string
	var extraAttrs string

	// Atributos do handler
	for _, a := range h.attrs {
		if a.Key == "bot" || a.Key == "account" {
			accountTag = fmt.Sprintf(" \033[35m[%s]\033[0m", a.Value.String())
		} else {
			extraAttrs += fmt.Sprintf(" %s=%v", a.Key, a.Value.Any())
		}
	}

	// Atributos do registro
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "bot" || a.Key == "account" {
			accountTag = fmt.Sprintf(" \033[35m[%s]\033[0m", a.Value.String())
		} else {
			extraAttrs += fmt.Sprintf(" %s=%v", a.Key, a.Value.Any())
		}
		return true
	})

	fmt.Fprintf(h.w, "%s %s%s %s%s\n", timeStr, levelTag, accountTag, r.Message, extraAttrs)
	return nil
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &ConsoleHandler{
		w:     h.w,
		level: h.level,
		attrs: newAttrs,
		group: h.group,
	}
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	return &ConsoleHandler{
		w:     h.w,
		level: h.level,
		attrs: h.attrs,
		group: name,
	}
}

// InitLogger configura o slog padrão
func InitLogger(level slog.Level) {
	loggerOnce.Do(func() {
		handler := NewConsoleHandler(os.Stdout, level)
		defaultLogger = slog.New(handler)
		slog.SetDefault(defaultLogger)
	})
}

// ForAccount devolve um logger com contexto de conta/bot
func ForAccount(accountID string) *slog.Logger {
	if defaultLogger == nil {
		InitLogger(slog.LevelInfo)
	}
	return defaultLogger.With("bot", accountID)
}
