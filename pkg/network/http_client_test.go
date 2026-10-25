package network

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClient_MethodsAndHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validar headers de stealth
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("User-Agent ausente")
		}
		if r.Header.Get("sec-ch-ua") == "" {
			t.Errorf("sec-ch-ua ausente")
		}

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"level": 100, "name": "Hero"})
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "action": body["action"]})
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"saved": true, "rev": 42})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer ts.Close()

	client := NewHTTPClient(ts.URL, 5*time.Second)
	ctx := context.Background()

	// 1. Testar GET
	resp, err := client.Get(ctx, "/api/state")
	if err != nil {
		t.Fatalf("GET falhou: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatalf("GET status não é sucesso: %d", resp.StatusCode)
	}
	var state struct {
		Level int    `json:"level"`
		Name  string `json:"name"`
	}
	if err := resp.Unmarshal(&state); err != nil {
		t.Fatalf("Unmarshal falhou: %v", err)
	}
	if state.Level != 100 || state.Name != "Hero" {
		t.Errorf("Dados inesperados: %+v", state)
	}

	// 2. Testar POST
	postResp, err := client.PostJSON(ctx, "/api/action", map[string]string{"action": "attack"})
	if err != nil {
		t.Fatalf("POST falhou: %v", err)
	}
	if postResp.StatusCode != http.StatusCreated {
		t.Errorf("Status POST inesperado: %d", postResp.StatusCode)
	}

	// 3. Testar PUT (Save)
	putResp, err := client.PutJSON(ctx, "/api/save", map[string]any{"gold": 5000})
	if err != nil {
		t.Fatalf("PUT falhou: %v", err)
	}
	if !putResp.IsSuccess() {
		t.Errorf("Status PUT inesperado: %d", putResp.StatusCode)
	}

	// 4. Testar DELETE
	delResp, err := client.Delete(ctx, "/api/session")
	if err != nil {
		t.Fatalf("DELETE falhou: %v", err)
	}
	if delResp.StatusCode != http.StatusNoContent {
		t.Errorf("Status DELETE inesperado: %d", delResp.StatusCode)
	}
}

func TestHTTPClient_AntiCheatRejection(t *testing.T) {
	resp429 := &Response{
		StatusCode: http.StatusTooManyRequests,
		Body:       []byte("Too Many Requests"),
	}
	if !resp429.IsAntiCheatRejection() {
		t.Errorf("Esperava 429 marcado como anti-cheat rejection")
	}

	respFastExp := &Response{
		StatusCode: http.StatusBadRequest,
		Body:       []byte(`{"erro": "experiência rápida demais para este personagem"}`),
	}
	if !respFastExp.IsAntiCheatRejection() {
		t.Errorf("Esperava 'experiência rápida demais' marcado como anti-cheat rejection")
	}

	respRecusado := &Response{
		StatusCode: http.StatusOK,
		Body:       []byte(`{"recusado": true, "motivo": "save malformado"}`),
	}
	if !respRecusado.IsAntiCheatRejection() {
		t.Errorf("Esperava 'recusado: true' marcado como anti-cheat rejection")
	}
}
