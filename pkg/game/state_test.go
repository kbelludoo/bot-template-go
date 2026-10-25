package game

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCanonicalStatePreservesUnknownFieldsAndLargeIntegers(t *testing.T) {
	var state PokeTrueState
	if err := json.Unmarshal([]byte(`{"nivel":5,"gold":10,"revision":9007199254740993,"unmodeled":{"skins":["rare"]}}`), &state); err != nil {
		t.Fatal(err)
	}
	if state.Nivel != 5 || state.Gold != 10 {
		t.Fatal("campos conhecidos não foram decodificados")
	}
	state.RawEstado["gold"] = 20
	envelope, err := json.Marshal(map[string]any{"estado": state.RawEstado})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"revision":9007199254740993`, `"skins":["rare"]`, `"gold":20`} {
		if !strings.Contains(string(envelope), expected) {
			t.Fatalf("campo perdido: %s", expected)
		}
	}
}
