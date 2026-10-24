package network

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func TestSocketClient_RoutingAndSend(t *testing.T) {
	var serverReceivedText string
	var serverReceivedBinary []byte
	var mu sync.Mutex
	_ = serverReceivedText
	_ = serverReceivedBinary

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Upgrader failed: %v", err)
			return
		}
		defer conn.Close()

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			mu.Lock()
			if mt == websocket.TextMessage {
				serverReceivedText = string(message)
				// Responder com mensagem tipada
				_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"battle_result","damage":150}`))
			} else if mt == websocket.BinaryMessage {
				serverReceivedBinary = message
				// Ecoar binário
				_ = conn.WriteMessage(websocket.BinaryMessage, append([]byte{0xFF}, message...))
			}
			mu.Unlock()
		}
	}))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http")

	cfg := DefaultSocketConfig(wsURL)
	cfg.PingInterval = 100 * time.Millisecond
	cfg.PongWait = 500 * time.Millisecond

	client := NewSocketClientWithConfig(cfg)

	battleResultChan := make(chan int, 1)
	client.On("battle_result", func(data []byte) {
		var res struct {
			Damage int `json:"damage"`
		}
		_ = json.Unmarshal(data, &res)
		battleResultChan <- res.Damage
	})

	binaryChan := make(chan []byte, 1)
	client.OnBinary(func(data []byte) {
		binaryChan <- data
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.ConnectWithContext(ctx); err != nil {
		t.Fatalf("Connect falhou: %v", err)
	}
	defer client.Close()

	if !client.IsConnected() {
		t.Fatalf("Socket deveria estar conectado")
	}

	// 1. Enviar JSON
	if err := client.SendJSON(map[string]string{"type": "attack", "target": "goblin"}); err != nil {
		t.Fatalf("SendJSON falhou: %v", err)
	}

	// Aguardar resposta roteada
	select {
	case dmg := <-battleResultChan:
		if dmg != 150 {
			t.Errorf("Esperava dano 150, obteve: %d", dmg)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Timeout aguardando battle_result")
	}

	// 2. Enviar Binário
	rawBytes := []byte{0x08, 0x96, 0x01}
	if err := client.SendBinary(rawBytes); err != nil {
		t.Fatalf("SendBinary falhou: %v", err)
	}

	select {
	case bin := <-binaryChan:
		if len(bin) != 4 || bin[0] != 0xFF {
			t.Errorf("Resposta binária inesperada: %v", bin)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Timeout aguardando resposta binária")
	}
}

func echoServer(t *testing.T, disconnectFirst bool) string {
	t.Helper()
	var connections atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		if connections.Add(1) == 1 && disconnectFirst {
			return
		}
		for {
			kind, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(kind, data); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)
	return "ws" + strings.TrimPrefix(srv.URL, "http")
}

func TestReconnectInsideDisconnectCallbackPreservesNewSession(t *testing.T) {
	url := echoServer(t, true)
	reconnected := make(chan error, 4)
	message := make(chan string, 4)
	var client *SocketClient
	client = NewSocketClient(url, nil, func(data []byte) { message <- string(data) }, func(error) { reconnected <- client.Connect() })
	t.Cleanup(client.Close)
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-reconnected:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("callback não reconectou")
	}
	if err := client.SendText("new-session"); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-message:
		if got != "new-session" {
			t.Fatalf("mensagem: %s", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("sessão anterior fechou a nova conexão")
	}
	if !client.IsConnected() {
		t.Fatal("nova sessão desconectada")
	}
}

func TestSocketConcurrentSendPingAndClose(t *testing.T) {
	client := NewSocketClient(echoServer(t, false), nil, nil, nil)
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	if err := client.Connect(); err == nil {
		t.Fatal("conexão duplicada aceita")
	}
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 10 {
				_ = client.SendJSON(map[string]int{"test": 1})
				_ = client.SendPing()
			}
		}()
	}
	client.Close()
	wg.Wait()
	client.Close()
	if client.IsConnected() {
		t.Fatal("Close manteve conexão ativa")
	}
	if err := client.SendText("closed"); err == nil {
		t.Fatal("envio após Close aceito")
	}
}

func TestCloseDuringHandshakeCancelsPendingConnection(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		conn, err := upgrader.Upgrade(w, r, nil)
		if err == nil {
			defer conn.Close()
			_, _, _ = conn.ReadMessage()
		}
	}))
	defer srv.Close()
	client := NewSocketClient("ws"+strings.TrimPrefix(srv.URL, "http"), nil, nil, nil)
	defer client.Close()
	done := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() { done <- client.ConnectContext(ctx) }()
	select {
	case <-entered:
	case <-ctx.Done():
		close(release)
		t.Fatal("handshake não iniciou")
	}
	client.Close()
	close(release)
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("conexão pendente sobreviveu ao Close")
		}
	case <-ctx.Done():
		t.Fatal("conexão não terminou")
	}
	if client.IsConnected() {
		t.Fatal("conexão cancelada foi instalada")
	}
}
