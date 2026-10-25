package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"
)

// Response representa a resposta encapsulada de uma requisição HTTP
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
}

// Unmarshal decodifica o corpo JSON para a struct de destino
func (r *Response) Unmarshal(target any) error {
	if len(r.Body) == 0 {
		return fmt.Errorf("corpo da resposta vazio")
	}
	return json.Unmarshal(r.Body, target)
}

// String retorna o corpo como string
func (r *Response) String() string {
	return string(r.Body)
}

// IsSuccess verifica se o status é 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsAntiCheatRejection checa se a resposta indica bloqueio ou detecção anti-cheat
func (r *Response) IsAntiCheatRejection() bool {
	if r.StatusCode == http.StatusTooManyRequests || r.StatusCode == http.StatusForbidden {
		return true
	}
	bodyLower := strings.ToLower(string(r.Body))
	triggers := []string{
		"experiência rápida demais",
		"pontos de atributo demais",
		"contadores voltaram para trás",
		"save malformado",
		"\"recusado\":true",
		"\"recusado\": true",
		"\"outraaba\":true",
		"\"outraaba\": true",
		"\"banido\":true",
		"\"banido\": true",
		"cf-turnstile",
		"challenge-form",
	}
	for _, t := range triggers {
		if strings.Contains(bodyLower, t) {
			return true
		}
	}
	return false
}

// HTTPClientConfig define os parâmetros de conexão do cliente HTTP
type HTTPClientConfig struct {
	BaseURL         string
	Timeout         time.Duration
	BindIP          string
	UserAgent       string
	EnableCookieJar bool
	MaxRetries      int
	MaxConcurrent   int // Quando > 0, limita requisições concorrentes da mesma conta (ex: 1 = fila serial)
	Headers         map[string]string
	Signer          *RequestSigner
}

// DefaultHTTPConfig devolve configurações padrão otimizadas para stealth e resiliência
func DefaultHTTPConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:         15 * time.Second,
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		EnableCookieJar: true,
		MaxRetries:      2,
		Headers: map[string]string{
			"Accept":             "application/json, text/plain, */*",
			"Accept-Language":    "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7",
			"sec-ch-ua":          "\"Google Chrome\";v=\"131\", \"Chromium\";v=\"131\", \"Not_A Brand\";v=\"24\"",
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": "\"Windows\"",
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-origin",
		},
	}
}

// HTTPClient cliente HTTP concorrente com suporte a multi-IP, cookies, assinatura HMAC e headers de stealth
type HTTPClient struct {
	mu          sync.RWMutex
	client      *http.Client
	baseURL     string
	bindIP      string
	headers     map[string]string
	maxRetries  int
	sem         chan struct{}
	signer      *RequestSigner
	shardParam  string
	shardHeader string
	shardValue  string
}

// NewHTTPClient inicializa o cliente HTTP com suporte opcional a IP secundário (Multi-IP)
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	cfg := DefaultHTTPConfig()
	cfg.BaseURL = baseURL
	if timeout > 0 {
		cfg.Timeout = timeout
	}
	return NewHTTPClientWithConfig(cfg)
}

// NewHTTPClientWithConfig inicializa o cliente com configuração detalhada
func NewHTTPClientWithConfig(cfg HTTPClientConfig) *HTTPClient {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	// Suporte a Multi-IP para rodar múltiplas contas na mesma VPS
	if cfg.BindIP != "" {
		localIP := net.ParseIP(cfg.BindIP)
		if localIP != nil {
			dialer := &net.Dialer{
				LocalAddr: &net.TCPAddr{IP: localIP},
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			transport.DialContext = dialer.DialContext
		}
	}

	var jar http.CookieJar
	if cfg.EnableCookieJar {
		jar, _ = cookiejar.New(nil)
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		Jar:       jar,
	}

	hdrs := make(map[string]string)
	for k, v := range cfg.Headers {
		hdrs[k] = v
	}
	hdrs["User-Agent"] = cfg.UserAgent

	var sem chan struct{}
	if cfg.MaxConcurrent > 0 {
		sem = make(chan struct{}, cfg.MaxConcurrent)
	}

	return &HTTPClient{
		client:     httpClient,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		bindIP:     cfg.BindIP,
		headers:    hdrs,
		maxRetries: cfg.MaxRetries,
		sem:        sem,
		signer:     cfg.Signer,
	}
}

// SetSigner define ou atualiza o assinador HMAC-SHA256 do cliente
func (c *HTTPClient) SetSigner(signer *RequestSigner) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.signer = signer
}

// SetShard configura roteamento de partição/shard dinâmico (ex: ?p=3 e X-Pokeidle-P: 3)
func (c *HTTPClient) SetShard(paramName, headerName, shardValue string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shardParam = paramName
	c.shardHeader = headerName
	c.shardValue = shardValue
}

// SetHeader define ou sobrescreve um header padrão
func (c *HTTPClient) SetHeader(k, v string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.headers[k] = v
}

// SetCookie adiciona cookie na sessão
func (c *HTTPClient) SetCookie(cookies string) {
	c.SetHeader("Cookie", cookies)
}

// GetHeader lê um header
func (c *HTTPClient) GetHeader(k string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.headers[k]
}

// Do executa uma requisição HTTP com headers de stealth, fila de concorrência e medição de tempo
func (c *HTTPClient) Do(ctx context.Context, req *http.Request) (*Response, error) {
	if c.sem != nil {
		select {
		case c.sem <- struct{}{}:
			defer func() { <-c.sem }()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	var bodyBytesForRetry []byte
	if req.Body != nil {
		var err error
		bodyBytesForRetry, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("falha ao ler corpo da requisição: %w", err)
		}
		_ = req.Body.Close()
	}

	c.mu.RLock()
	for k, v := range c.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	if c.shardHeader != "" && c.shardValue != "" {
		req.Header.Set(c.shardHeader, c.shardValue)
	}
	if c.shardParam != "" && c.shardValue != "" {
		q := req.URL.Query()
		if q.Get(c.shardParam) == "" {
			q.Set(c.shardParam, c.shardValue)
			req.URL.RawQuery = q.Encode()
		}
	}
	signer := c.signer
	c.mu.RUnlock()

	if signer != nil {
		signer.SignRequest(req, bodyBytesForRetry)
	}

	start := time.Now()
	var resp *http.Response
	var err error

	attempts := c.maxRetries + 1
	for attempt := 0; attempt < attempts; attempt++ {
		reqWithCtx := req.Clone(ctx)
		if bodyBytesForRetry != nil {
			reqWithCtx.Body = io.NopCloser(bytes.NewReader(bodyBytesForRetry))
			reqWithCtx.ContentLength = int64(len(bodyBytesForRetry))
		}
		resp, err = c.client.Do(reqWithCtx)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		// Se for cancelado pelo contexto, abortar retries
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if attempt < attempts-1 {
			// Backoff exponencial com jitter
			jitter := time.Duration(rand.Intn(200)) * time.Millisecond
			sleepDuration := time.Duration(1<<attempt)*300*time.Millisecond + jitter
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(sleepDuration):
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("requisição falhou após retries (%s %s): %w", req.Method, req.URL.String(), err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler corpo da resposta: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       bodyBytes,
		Duration:   time.Since(start),
	}, nil
}

// Get executa uma requisição GET
func (c *HTTPClient) Get(ctx context.Context, endpoint string) (*Response, error) {
	url := c.resolveURL(endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

// PostJSON envia payload JSON via POST
func (c *HTTPClient) PostJSON(ctx context.Context, endpoint string, reqBody any) (*Response, error) {
	return c.sendJSON(ctx, http.MethodPost, endpoint, reqBody)
}

// PutJSON envia payload JSON via PUT (essencial para salvamento /api/save em nuvem)
func (c *HTTPClient) PutJSON(ctx context.Context, endpoint string, reqBody any) (*Response, error) {
	return c.sendJSON(ctx, http.MethodPut, endpoint, reqBody)
}

// PatchJSON envia payload JSON via PATCH
func (c *HTTPClient) PatchJSON(ctx context.Context, endpoint string, reqBody any) (*Response, error) {
	return c.sendJSON(ctx, http.MethodPatch, endpoint, reqBody)
}

// Delete executa requisição DELETE
func (c *HTTPClient) Delete(ctx context.Context, endpoint string) (*Response, error) {
	url := c.resolveURL(endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(ctx, req)
}

func (c *HTTPClient) sendJSON(ctx context.Context, method, endpoint string, reqBody any) (*Response, error) {
	var bodyReader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar json: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.resolveURL(endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.Do(ctx, req)
}

func (c *HTTPClient) resolveURL(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	return c.baseURL + endpoint
}
