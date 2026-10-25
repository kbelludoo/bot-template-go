package network

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RequestSigner assina requisições HTTP com HMAC-SHA256 e injeta Idempotency-Key (padrão PokePixel)
type RequestSigner struct {
	mu              sync.RWMutex
	signingKey      string
	timestampHeader string
	signatureHeader string
}

// NewRequestSigner cria um assinador HMAC-SHA256 para autenticação de integridade de payload
func NewRequestSigner(signingKey string) *RequestSigner {
	return &RequestSigner{
		signingKey:      signingKey,
		timestampHeader: "X-Request-Timestamp",
		signatureHeader: "X-Request-Signature",
	}
}

// SetKey atualiza a chave de assinatura obtida no login
func (s *RequestSigner) SetKey(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.signingKey = key
}

// GetKey retorna a chave de assinatura atual
func (s *RequestSigner) GetKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.signingKey
}

// ComputeSignature calcula HMAC-SHA256(key, METHOD + "\n" + PathWithoutQuery + "\n" + TimestampMs + "\n" + SHA256Hex(body))
func (s *RequestSigner) ComputeSignature(method, pathWithoutQuery, timestampMs string, body []byte) string {
	s.mu.RLock()
	key := s.signingKey
	s.mu.RUnlock()
	if key == "" {
		return ""
	}

	bodySum := sha256.Sum256(body)
	bodyHex := hex.EncodeToString(bodySum[:])

	canonical := strings.ToUpper(method) + "\n" + pathWithoutQuery + "\n" + timestampMs + "\n" + bodyHex
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil))
}

// SignRequest aplica os cabeçalhos X-Request-Timestamp, X-Request-Signature e Idempotency-Key na requisição
func (s *RequestSigner) SignRequest(req *http.Request, body []byte) {
	if s == nil {
		return
	}
	s.mu.RLock()
	key := s.signingKey
	tsHeader := s.timestampHeader
	sigHeader := s.signatureHeader
	s.mu.RUnlock()

	if key == "" {
		return
	}

	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	path := req.URL.Path
	if path == "" {
		path = "/"
	}

	sig := s.ComputeSignature(req.Method, path, ts, body)
	req.Header.Set(tsHeader, ts)
	req.Header.Set(sigHeader, sig)

	// Injeta Idempotency-Key em mutações se ainda não definido
	if (req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch) && req.Header.Get("Idempotency-Key") == "" {
		req.Header.Set("Idempotency-Key", GenerateUUIDv4())
	}
}

// GenerateUUIDv4 gera um UUID v4 criptograficamente seguro
func GenerateUUIDv4() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// ExtractJWTClaim decodifica o payload de um JWT (sem verificar assinatura) e extrai uma claim string (ex: "sub")
func ExtractJWTClaim(jwtToken, claim string) string {
	parts := strings.Split(strings.TrimSpace(jwtToken), ".")
	if len(parts) < 2 {
		return ""
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return ""
	}
	if val, ok := claims[claim]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// DJB2HashInt32 executa o hash DJB2 com aritmética de int32 com sinal (idêntico ao bitwise `((h << 5) + h) + c | 0` do JavaScript)
func DJB2HashInt32(input string) int32 {
	var h int32 = 5381
	for i := 0; i < len(input); i++ {
		h = ((h << 5) + h) + int32(input[i])
	}
	return h
}

// ComputeDJB2ShardPath calcula a rota de shard WebSocket via DJB2 sobre o JWT sub (padrão PokeIdleWorld: 88 shards /ws0../ws87)
func ComputeDJB2ShardPath(jwtOrSub string, numShards int, prefix string) string {
	if numShards <= 0 {
		numShards = 88
	}
	if prefix == "" {
		prefix = "/ws"
	}
	sub := ExtractJWTClaim(jwtOrSub, "sub")
	if sub == "" {
		sub = jwtOrSub
	}
	h := int(DJB2HashInt32(sub))
	if h < 0 {
		h = -h
	}
	shard := h % numShards
	return fmt.Sprintf("%s%d", prefix, shard)
}

// GenerateDeviceFingerprint gera um deviceFingerprint determinístico por conta (padrão Baiak Idle: <32hex>-<djb2hex>)
func GenerateDeviceFingerprint(accountID, userAgent string) string {
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	}
	seed := sha256.Sum256([]byte("device-fp:" + accountID))
	randomHex32 := hex.EncodeToString(seed[:16])

	fpProfile := fmt.Sprintf("%s|1920|1080|24|America/Sao_Paulo|pt-BR", userAgent)
	h := uint32(DJB2HashInt32(fpProfile))
	return fmt.Sprintf("%s-%08x", randomHex32, h)
}
