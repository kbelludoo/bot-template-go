package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MessageHandler func(data []byte)
type DisconnectHandler func(err error)

// SocketClient mantém cada leitor associado à sua própria conexão.
type SocketClient struct {
	url          string
	headers      http.Header
	conn         *websocket.Conn
	writeMu      sync.Mutex
	dialMu       sync.Mutex
	mu           sync.RWMutex
	generation   uint64
	onMessage    MessageHandler
	onDisconnect DisconnectHandler
}

func NewSocketClient(url string, headers http.Header, onMsg MessageHandler, onDisc DisconnectHandler) *SocketClient {
	headers = headers.Clone()
	if headers == nil {
		headers = make(http.Header)
	}
	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	}
	return &SocketClient{url: url, headers: headers, onMessage: onMsg, onDisconnect: onDisc}
}

func (s *SocketClient) Connect() error {
	return s.ConnectContext(context.Background())
}

func (s *SocketClient) ConnectContext(ctx context.Context) error {
	s.dialMu.Lock()
	defer s.dialMu.Unlock()
	s.mu.RLock()
	connected := s.conn != nil
	generation := s.generation
	s.mu.RUnlock()
	if connected {
		return fmt.Errorf("socket já conectado")
	}
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, resp, err := dialer.DialContext(ctx, s.url, s.headers)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return fmt.Errorf("websocket dial failed: %w", err)
	}
	conn.SetReadLimit(10 * 1024 * 1024)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})
	s.mu.Lock()
	if s.generation != generation {
		s.mu.Unlock()
		_ = conn.Close()
		return fmt.Errorf("conexão cancelada por Close")
	}
	s.conn = conn
	s.mu.Unlock()
	done := make(chan struct{})
	go s.readLoop(conn, done)
	go s.keepAlive(conn, done)
	return nil
}

func (s *SocketClient) readLoop(conn *websocket.Conn, done chan struct{}) {
	defer close(done)
	defer func() {
		if recovered := recover(); recovered != nil {
			s.disconnect(conn, fmt.Errorf("panic no handler websocket: %v", recovered))
		}
	}()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			s.disconnect(conn, err)
			return
		}
		if s.onMessage != nil {
			s.onMessage(message)
		}
	}
}

func (s *SocketClient) keepAlive(conn *websocket.Conn, done <-chan struct{}) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				s.disconnect(conn, err)
				return
			}
		}
	}
}

// disconnect limpa apenas a sessão que falhou, antes de permitir reconexão no callback.
func (s *SocketClient) disconnect(conn *websocket.Conn, err error) {
	s.mu.Lock()
	current := s.conn == conn
	if current {
		s.conn = nil
	}
	s.mu.Unlock()
	_ = conn.Close()
	if current && s.onDisconnect != nil {
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					log.Printf("panic no handler de desconexão websocket: %v", recovered)
				}
			}()
			s.onDisconnect(err)
		}()
	}
}

func (s *SocketClient) SendJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.SendText(string(data))
}

func (s *SocketClient) SendText(text string) error {
	s.writeMu.Lock()
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		s.writeMu.Unlock()
		return fmt.Errorf("socket não conectado")
	}
	err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err == nil {
		err = conn.WriteMessage(websocket.TextMessage, []byte(text))
	}
	s.writeMu.Unlock()
	if err != nil {
		s.disconnect(conn, err)
	}
	return err
}

func (s *SocketClient) SendPing() error {
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		return nil
	}
	err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
	if err != nil {
		s.disconnect(conn, err)
	}
	return err
}

func (s *SocketClient) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.conn != nil
}

func (s *SocketClient) Close() {
	s.mu.Lock()
	conn := s.conn
	s.conn = nil
	s.generation++
	s.mu.Unlock()
	if conn != nil {
		// WriteControl permite Close concorrente com o escritor de dados.
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		_ = conn.Close()
	}
}
