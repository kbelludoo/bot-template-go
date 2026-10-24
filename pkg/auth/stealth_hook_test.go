package auth

import (
	"strings"
	"testing"
)

func TestStealthHookAndLaudo(t *testing.T) {
	s := StealthInitScript()
	if !strings.Contains(s, "navigator, 'webdriver'") {
		t.Fatalf("StealthInitScript incompleto")
	}

	hook := TurnstileInterceptScript("/auth/world-ticket")
	if !strings.Contains(hook, "__INTERCEPTED_TURNSTILE_TOKEN") || !strings.Contains(hook, "DummyWebSocket") {
		t.Fatalf("TurnstileInterceptScript incompleto")
	}

	laudo := BuildCleanAntiBotLaudo()
	if laudo["v"] != "detector-2" || laudo["pontos"] != 0 {
		t.Fatalf("laudo inválido: %+v", laudo)
	}
	sinais, ok := laudo["sinais"].(map[string]string)
	if !ok || len(sinais) != 14 || sinais["wd"] != "nao" {
		t.Fatalf("sinais do laudo inválidos: %+v", sinais)
	}
}
