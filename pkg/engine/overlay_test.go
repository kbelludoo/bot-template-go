package engine

import (
	"encoding/json"
	"testing"
)

func TestStateOverlay_PreservesUnknownFields(t *testing.T) {
	rawJSON := []byte(`{
		"nivel": 25,
		"ouro": 1500,
		"rev": 10,
		"conquistas": ["first_blood", "dragon_slayer"],
		"montaria": {"id": 4, "nome": "Lobo", "speed": 1.2},
		"skins": {"ativa": "cyber_knight"},
		"gemas": [101, 102, 105],
		"totais": {
			"abates": 450,
			"tempoJogo": 12000
		}
	}`)

	var raw map[string]any
	if err := json.Unmarshal(rawJSON, &raw); err != nil {
		t.Fatalf("Unmarshal raw failed: %v", err)
	}

	overlay := NewStateOverlay(raw)

	overlay.Set("ouro", 2000)
	overlay.Set("totais.abates", 455)
	overlay.UpdateRev(11)

	env := overlay.Envelope("estado")
	estado, ok := env["estado"].(map[string]any)
	if !ok {
		t.Fatalf("Envelope 'estado' inválido")
	}

	if estado["ouro"] != 2000 {
		t.Errorf("Ouro não foi atualizado: %v", estado["ouro"])
	}
	totais := estado["totais"].(map[string]any)
	if totais["abates"] != 455 {
		t.Errorf("Abates não foi atualizado: %v", totais["abates"])
	}
	if estado["rev"] != int64(11) {
		t.Errorf("Rev não foi atualizado: %v", estado["rev"])
	}

	if _, has := estado["conquistas"]; !has {
		t.Errorf("Campo 'conquistas' foi perdido no overlay!")
	}
	if _, has := estado["montaria"]; !has {
		t.Errorf("Campo 'montaria' foi perdido no overlay!")
	}
	if _, has := estado["skins"]; !has {
		t.Errorf("Campo 'skins' foi perdido no overlay!")
	}
	if _, has := estado["gemas"]; !has {
		t.Errorf("Campo 'gemas' foi perdido no overlay!")
	}
}

func TestStateOverlay_ContractReconciliation409(t *testing.T) {
	if !IsStaleContractConflict(409, `{"error":"SAVE_CONTRACT_STALE"}`) {
		t.Fatal("deveria reconhecer 409 SAVE_CONTRACT_STALE")
	}

	overlay := NewStateOverlay(map[string]any{"gold": 100})
	overlay.SetSaveContract(5, "v1.0.0", []string{"inventory", "depot"})

	err := overlay.ReconcileStaleSave(func() ([]byte, error) {
		return []byte(`{
			"saveVersion": 9,
			"clientVersion": "v1.0.2",
			"saveOwnedFields": ["inventory", "depot", "quests"],
			"save": {"gold": 500, "gems": 10}
		}`), nil
	}, func(so *StateOverlay) {
		so.Set("gold", 650)
	})
	if err != nil {
		t.Fatalf("ReconcileStaleSave falhou: %v", err)
	}

	if overlay.SaveVersion() != 9 {
		t.Fatalf("esperava saveVersion 9, obteve %d", overlay.SaveVersion())
	}
	cEnv := overlay.ContractEnvelope("save")
	saveMap := cEnv["save"].(map[string]any)
	if saveMap["gold"] != 650 || saveMap["gems"] != float64(10) {
		t.Fatalf("merge após reconciliação incorreto: %+v", saveMap)
	}
	omitted := cEnv["omitted"].([]string)
	if len(omitted) != 3 {
		t.Fatalf("esperava 3 campos omitidos, obteve %v", omitted)
	}
}

func TestStateOverlay_Monotonicity(t *testing.T) {
	raw := map[string]any{
		"totais": map[string]any{
			"abates": 100,
		},
	}
	overlay := NewStateOverlay(raw)

	if err := overlay.CheckMonotonicity("totais.abates", 105); err != nil {
		t.Errorf("Esperava sucesso no incremento, obteve: %v", err)
	}
	if err := overlay.CheckMonotonicity("totais.abates", 110); err != nil {
		t.Errorf("Esperava sucesso no incremento, obteve: %v", err)
	}
	if err := overlay.CheckMonotonicity("totais.abates", 108); err == nil {
		t.Errorf("Esperava erro de monotonicidade ao diminuir abates de 110 para 108")
	}
}

func TestStateOverlay_AttributeBudget(t *testing.T) {
	if err := VerifyAttributeBudget(20, 7, 10, 3); err != nil {
		t.Errorf("Esperava orçamento válido, obteve erro: %v", err)
	}
	if err := VerifyAttributeBudget(25, 5, 10, 3); err == nil {
		t.Errorf("Esperava falha ao exceder orçamento de atributos")
	}
}
