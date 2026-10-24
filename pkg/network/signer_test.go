package network

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestSignerAndSharding(t *testing.T) {
	signer := NewRequestSigner("super-secret-signing-key")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Request-Timestamp") == "" {
			t.Errorf("esperava X-Request-Timestamp")
		}
		if r.Header.Get("X-Request-Signature") == "" {
			t.Errorf("esperava X-Request-Signature")
		}
		if r.Header.Get("Idempotency-Key") == "" {
			t.Errorf("esperava Idempotency-Key em POST")
		}
		if r.Header.Get("X-Pokeidle-P") != "3" {
			t.Errorf("esperava X-Pokeidle-P=3, obteve %s", r.Header.Get("X-Pokeidle-P"))
		}
		if r.URL.Query().Get("p") != "3" {
			t.Errorf("esperava query p=3, obteve %s", r.URL.Query().Get("p"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	cfg := DefaultHTTPConfig()
	cfg.BaseURL = srv.URL
	cfg.MaxConcurrent = 1
	cfg.Signer = signer
	client := NewHTTPClientWithConfig(cfg)
	client.SetShard("p", "X-Pokeidle-P", "3")

	resp, err := client.PostJSON(context.Background(), "/api/save", map[string]any{"gold": 100})
	if err != nil {
		t.Fatalf("erro na requisição: %v", err)
	}
	if !resp.IsSuccess() {
		t.Fatalf("status inesperado: %d", resp.StatusCode)
	}
}

func TestDJB2ShardAndDeviceFingerprint(t *testing.T) {
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user_987654321"}`))
	fakeJWT := "eyJhbGciOiJIUzI1NiJ9." + payload + ".sig"

	if sub := ExtractJWTClaim(fakeJWT, "sub"); sub != "user_987654321" {
		t.Fatalf("esperava user_987654321, obteve %s", sub)
	}

	shardPath := ComputeDJB2ShardPath(fakeJWT, 88, "/ws")
	if !strings.HasPrefix(shardPath, "/ws") {
		t.Fatalf("rota de shard inválida: %s", shardPath)
	}

	fp1 := GenerateDeviceFingerprint("acc_1", "")
	fp2 := GenerateDeviceFingerprint("acc_1", "")
	if fp1 != fp2 || len(fp1) != 41 {
		t.Fatalf("fingerprint determinístico inválido: %s", fp1)
	}
}
