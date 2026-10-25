package example

import (
	"context"
	"fmt"
	"time"

	"bot-template-go/pkg/engine"
	"bot-template-go/pkg/network"
)

// Client comunica-se com a API oficial do Example Idle
type Client struct {
	http *network.HTTPClient
	ws   *network.SocketClient
	cb   *engine.CircuitBreaker
}

func NewClient(baseURL, bindIP string, cb *engine.CircuitBreaker) *Client {
	cfg := network.DefaultHTTPConfig()
	cfg.BaseURL = baseURL
	cfg.BindIP = bindIP
	cfg.Timeout = 15 * time.Second

	return &Client{
		http: network.NewHTTPClientWithConfig(cfg),
		cb:   cb,
	}
}

// FetchState carrega o estado canônico do personagem
func (c *Client) FetchState(ctx context.Context) (*network.Response, error) {
	resp, err := c.http.Get(ctx, "/api/eu")
	if err != nil {
		c.cb.RecordFailure(err.Error())
		return nil, err
	}
	if resp.IsAntiCheatRejection() {
		c.cb.Trip(fmt.Sprintf("Anti-Cheat Rejection HTTP %d: %s", resp.StatusCode, resp.String()))
		return nil, fmt.Errorf("ação rejeitada pelo anti-cheat: %s", resp.String())
	}
	c.cb.RecordSuccess()
	return resp, nil
}

// SaveProgress salva o estado com overlay preservando chaves desconhecidas
func (c *Client) SaveProgress(ctx context.Context, overlay *engine.StateOverlay) (*network.Response, error) {
	if !c.cb.AllowAction() {
		return nil, fmt.Errorf("circuit breaker desarmado, save abortado para evitar ban")
	}

	payload := overlay.Envelope("estado")
	resp, err := c.http.PutJSON(ctx, "/api/save", payload)
	if err != nil {
		c.cb.RecordFailure(err.Error())
		return nil, err
	}
	if resp.IsAntiCheatRejection() {
		c.cb.Trip(fmt.Sprintf("Recusa em PUT /api/save: %s", resp.String()))
		return nil, fmt.Errorf("save rejeitado: %s", resp.String())
	}
	c.cb.RecordSuccess()
	return resp, nil
}
