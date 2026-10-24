package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type MessageHandler func(data []byte)
type DisconnectHandler func(err error)

// SocketConfig encapsula todas as configurações do cliente WebSocket
type SocketConfig struct {
	URL            string
	Headers        http.Header
	BindIP         string
	PingInterval   time.Duration
	PongWait       time.Duration
	WriteWait      time.Duration
	MaxMessageSize int64
	AutoReconnect  bool
	MaxBackoff     time.Duration
}

// DefaultSocketConfig devolve configurações resilientes com keepalive
func DefaultSocketConfig(url string) SocketConfig {
	headers := make(http.Header)
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	headers.Set("Origin", "https://localhost")

	return SocketConfig{
		URL:            url,
		Headers:        headers,
		PingInterval:   20 * time.Second,
		PongWait:       60 * time.Second,
		WriteWait:      10 * time.Second,
		MaxMessageSize: 10 * 1024 * 1024, // 10MB
		AutoReconnect:  true,
		MaxBackoff:     30 * time.Second,
	}
}

// SocketClient gerencia conexão WebSocket com escrita thread-safe, geração de sessão, roteamento de mensagens e keepalive
type SocketClient struct {
	cfg        SocketConfig
	conn       *websocket.Conn
	writeMu    sync.Mutex
	dialMu     sync.Mutex
	mu         sync.RWMutex
	generation uint64
	lastPingAt time.Time
	lastPongAt time.Time
	latency    time.Duration

	// Handlers globais
	onMessage    MessageHandler
	onDisconnect DisconnectHandler
	onConnect    func()

	// Roteador de mensagens por tipo JSON (ex: {"type": "combat"} -> chama handler "combat")
	routesMu      sync.RWMutex
	routes        map[string]MessageHandler
	binaryHandler MessageHandler

	closeChan chan struct{}
}

// NewSocketClient inicializa um cliente com opções básicas
func NewSocketClient(url string, headers http.Header, onMsg MessageHandler, onDisc DisconnectHandler) *SocketClient {
	cfg := DefaultSocketConfig(url)
	if headers != nil {
		for k, v := range headers {
			cfg.Headers[k] = v
		}
	}
	sc := NewSocketClientWithConfig(cfg)
	sc.onMessage = onMsg
	sc.onDisconnect = onDisc
	return sc
}

// NewSocketClientWithConfig cria o cliente WebSocket completo
func NewSocketClientWithConfig(cfg SocketConfig) *SocketClient {
	if cfg.Headers == nil {
		cfg.Headers = make(http.Header)
	}
	if cfg.Headers.Get("User-Agent") == "" {
		cfg.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	}
	if cfg.PingInterval <= 0 {
		cfg.PingInterval = 20 * time.Second
	}
	if cfg.PongWait <= 0 {
		cfg.PongWait = 60 * time.Second
	}
	if cfg.WriteWait <= 0 {
		cfg.WriteWait = 10 * time.Second
	}
	if cfg.MaxMessageSize <= 0 {
		cfg.MaxMessageSize = 10 * 1024 * 1024
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 30 * time.Second
	}

	return &SocketClient{
		cfg:       cfg,
		routes:    make(map[string]MessageHandler),
		closeChan: make(chan struct{}),
	}
}

// On registra um handler para um tipo específico de mensagem JSON (campo "type" ou "t")
func (s *SocketClient) On(messageType string, handler MessageHandler) {
	s.routesMu.Lock()
	defer s.routesMu.Unlock()
	s.routes[messageType] = handler
}

// OnBinary registra um handler para frames binários (Protobuf, MessagePack, Colyseus)
func (s *SocketClient) OnBinary(handler MessageHandler) {
	s.routesMu.Lock()
	defer s.routesMu.Unlock()
	s.binaryHandler = handler
}

// OnConnect registra callback de conexão estabelecida
func (s *SocketClient) OnConnect(cb func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onConnect = cb
}

// SetURL permite atualizar a URL de conexão antes de discar
func (s *SocketClient) SetURL(u string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.URL = u
}

// Connect conecta ao WebSocket de forma síncrona
func (s *SocketClient) Connect() error {
	return s.ConnectContext(context.Background())
}

// ConnectWithContext alias de ConnectContext
func (s *SocketClient) ConnectWithContext(ctx context.Context) error {
	return s.ConnectContext(ctx)
}

// ConnectContext conecta com proteção de geração de sessão, suporte a cancelamento de contexto e bindIP
func (s *SocketClient) ConnectContext(ctx context.Context) error {
	s.dialMu.Lock()
	defer s.dialMu.Unlock()

	s.mu.RLock()
	connected := s.conn != nil
	generation := s.generation
	wsURL := s.cfg.URL
	headers := s.cfg.Headers.Clone()
	bindIP := s.cfg.BindIP
	s.mu.RUnlock()

	if connected {
		return fmt.Errorf("socket já conectado")
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}

	// Suporte Multi-IP para rotação/isolamento de IP por conta
	if bindIP != "" {
		localIP := net.ParseIP(bindIP)
		if localIP != nil {
			netDialer := &net.Dialer{
				LocalAddr: &net.TCPAddr{IP: localIP},
				Timeout:   10 * time.Second,
			}
			dialer.NetDialContext = netDialer.DialContext
		}
	}

	conn, resp, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return fmt.Errorf("websocket dial falhou (%s): %w", wsURL, err)
	}

	conn.SetReadLimit(s.cfg.MaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(s.cfg.PongWait))
	conn.SetPongHandler(func(string) error {
		now := time.Now()
		s.mu.Lock()
		s.lastPongAt = now
		if !s.lastPingAt.IsZero() {
			s.latency = now.Sub(s.lastPingAt)
		}
		s.mu.Unlock()
		return conn.SetReadDeadline(now.Add(s.cfg.PongWait))
	})

	s.mu.Lock()
	if s.generation != generation {
		s.mu.Unlock()
		_ = conn.Close()
		return fmt.Errorf("conexão cancelada por Close")
	}
	s.conn = conn
	s.closeChan = make(chan struct{})
	onConn := s.onConnect
	s.mu.Unlock()

	if onConn != nil {
		go onConn()
	}

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
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			s.disconnect(conn, err)
			return
		}

		_ = conn.SetReadDeadline(time.Now().Add(s.cfg.PongWait))

		if msgType == websocket.BinaryMessage {
			s.routesMu.RLock()
			bHandler := s.binaryHandler
			s.routesMu.RUnlock()
			if bHandler != nil {
				bHandler(message)
			} else if s.onMessage != nil {
				s.onMessage(message)
			}
			continue
		}

		dispatched := s.dispatchByType(message)
		if !dispatched && s.onMessage != nil {
			s.onMessage(message)
		}
	}
}

func (s *SocketClient) dispatchByType(data []byte) bool {
	s.routesMu.RLock()
	hasRoutes := len(s.routes) > 0
	s.routesMu.RUnlock()

	if !hasRoutes {
		return false
	}

	var envelope struct {
		Type   string `json:"type"`
		T      string `json:"t"`
		Action string `json:"action"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return false
	}

	routeKey := envelope.Type
	if routeKey == "" {
		routeKey = envelope.T
	}
	if routeKey == "" {
		routeKey = envelope.Action
	}
	if routeKey == "" {
		return false
	}

	s.routesMu.RLock()
	handler, exists := s.routes[routeKey]
	s.routesMu.RUnlock()

	if exists && handler != nil {
		handler(data)
		return true
	}

	return false
}

func (s *SocketClient) keepAlive(conn *websocket.Conn, done <-chan struct{}) {
	ticker := time.NewTicker(s.cfg.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			s.mu.Lock()
			s.lastPingAt = time.Now()
			s.mu.Unlock()
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
	var ch chan struct{}
	if current {
		s.conn = nil
		ch = s.closeChan
	}
	s.mu.Unlock()
	_ = conn.Close()
	if ch != nil {
		select {
		case <-ch:
		default:
			close(ch)
		}
	}
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

// SendJSON envia payload serializado em JSON com proteção de concorrência
func (s *SocketClient) SendJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("erro ao serializar json: %w", err)
	}
	return s.SendText(string(data))
}

// SendText envia mensagem de texto com proteção de concorrência
func (s *SocketClient) SendText(text string) error {
	s.writeMu.Lock()
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		s.writeMu.Unlock()
		return fmt.Errorf("socket não conectado")
	}
	err := conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteWait))
	if err == nil {
		err = conn.WriteMessage(websocket.TextMessage, []byte(text))
	}
	s.writeMu.Unlock()
	if err != nil {
		s.disconnect(conn, err)
	}
	return err
}

// SendBinary envia pacote binário (Protobuf / Byte frame) com proteção de concorrência
func (s *SocketClient) SendBinary(data []byte) error {
	s.writeMu.Lock()
	s.mu.RLock()
	conn := s.conn
	s.mu.RUnlock()
	if conn == nil {
		s.writeMu.Unlock()
		return fmt.Errorf("socket não conectado")
	}
	err := conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteWait))
	if err == nil {
		err = conn.WriteMessage(websocket.BinaryMessage, data)
	}
	s.writeMu.Unlock()
	if err != nil {
		s.disconnect(conn, err)
	}
	return err
}

// SendPing envia quadro de controle Ping
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

// IsConnected informa o estado atual da conexão
func (s *SocketClient) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.conn != nil
}

// Latency devolve o último ping medido
func (s *SocketClient) Latency() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latency
}

// RunWithAutoReconnect mantém a conexão aberta indefinidamente com reconexão automática
func (s *SocketClient) RunWithAutoReconnect(ctx context.Context) error {
	backoff := 1 * time.Second
	for {
		select {
		case <-ctx.Done():
			s.Close()
			return ctx.Err()
		default:
		}

		err := s.ConnectContext(ctx)
		if err == nil {
			backoff = 1 * time.Second
			s.mu.RLock()
			ch := s.closeChan
			s.mu.RUnlock()
			select {
			case <-ctx.Done():
				s.Close()
				return ctx.Err()
			case <-ch:
			}
		}

		if !s.cfg.AutoReconnect {
			return err
		}

		jitter := time.Duration(rand.Intn(500)) * time.Millisecond
		sleepTime := backoff + jitter
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepTime):
			backoff = minDuration(backoff*2, s.cfg.MaxBackoff)
		}
	}
}

// Close fecha a conexão limpa e invalida handshakes pendentes da geração anterior
func (s *SocketClient) Close() {
	s.mu.Lock()
	conn := s.conn
	s.conn = nil
	s.generation++
	ch := s.closeChan
	s.mu.Unlock()

	if conn != nil {
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		_ = conn.Close()
	}
	if ch != nil {
		select {
		case <-ch:
		default:
			close(ch)
		}
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
