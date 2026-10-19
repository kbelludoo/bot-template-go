package game

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"bot-template-go/pkg/engine"
)

// PokeTrueClient gerencia todas as chamadas HTTP à API do PokeTrue
type PokeTrueClient struct {
	baseURL    string
	token      string
	windowID   string
	httpClient *http.Client
	cb         *engine.CircuitBreaker
}

// Gera um ID de janela único como no navegador (x-janela)
func generateWindowID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b) + fmt.Sprintf("%x", time.Now().UnixMilli())
}

// NewPokeTrueClient cria o cliente com headers de browser moderno e suporte a multi-IP
func NewPokeTrueClient(token string, bindIP string, cb *engine.CircuitBreaker) *PokeTrueClient {
	dialer := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if bindIP != "" {
		if ip := net.ParseIP(bindIP); ip != nil {
			dialer.LocalAddr = &net.TCPAddr{IP: ip}
		}
	}

	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     60 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &PokeTrueClient{
		baseURL:  "https://poketrue.net/api/",
		token:    token,
		windowID: generateWindowID(),
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   20 * time.Second,
		},
		cb: cb,
	}
}

func (c *PokeTrueClient) SetToken(t string) {
	c.token = t
}

func (c *PokeTrueClient) GetToken() string {
	return c.token
}

// DoRequest executa a requisição HTTP com validação no Circuit Breaker
func (c *PokeTrueClient) DoRequest(ctx context.Context, method, endpoint string, body any, target any) error {
	if c.cb != nil && !c.cb.AllowAction() {
		return fmt.Errorf("circuit breaker desarmado (resfriamento anti-cheat ativo)")
	}

	var bodyReader io.Reader
	var rawBody []byte
	if body != nil {
		var err error
		rawBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("erro ao serializar payload: %w", err)
		}
		bodyReader = bytes.NewReader(rawBody)
	}

	url := c.baseURL + strings.TrimPrefix(endpoint, "/")
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Origin", "https://poketrue.net")
	req.Header.Set("Referer", "https://poketrue.net/play")
	// Não enviar x-janela aleatório: o servidor do PokeTrue vincula a sessão à janela do login
	// Enviar x-janela divergente disparava HTTP 401 "sessao_substituida". Omitindo, o servidor aceita.

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("falha de rede (%s): %w", endpoint, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro lendo resposta: %w", err)
	}

	// Verificação anti-cheat obrigatória
	if c.cb != nil {
		if isRej, reason := engine.IsAntiCheatRejection(resp.StatusCode, string(respBytes)); isRej {
			c.cb.Trip(fmt.Sprintf("%s (endpoint: %s)", reason, endpoint))
			return fmt.Errorf("anti-cheat disparado: %s", reason)
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		if c.cb != nil {
			c.cb.Trip("Rate limit (HTTP 429)")
		}
		return fmt.Errorf("HTTP 429: rate limit atingido")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d em %s: %s", resp.StatusCode, endpoint, string(respBytes))
	}

	if c.cb != nil {
		c.cb.RecordSuccess()
	}

	if target != nil && len(respBytes) > 0 {
		if err := json.Unmarshal(respBytes, target); err != nil {
			return fmt.Errorf("falha ao interpretar JSON (%s): %w", endpoint, err)
		}
	}

	return nil
}

type LoginResponse struct {
	Token         string `json:"token"`
	User          string `json:"user"`
	PrecisaCodigo bool   `json:"precisa_codigo"`
	Erro          string `json:"erro"`
}

// Login realiza a autenticação via usuário e senha
func (c *PokeTrueClient) Login(ctx context.Context, user, pass, captchaToken string) (*LoginResponse, error) {
	payload := map[string]any{
		"user": user,
		"pass": pass,
		"disp": c.windowID,
	}
	if captchaToken != "" {
		payload["captcha"] = captchaToken
	}

	var resp LoginResponse
	err := c.DoRequest(ctx, http.MethodPost, "login", payload, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Token != "" {
		c.token = resp.Token
	}
	return &resp, nil
}

type MeResponse struct {
	User          string `json:"user"`
	Diamonds      int    `json:"diamonds"`
	DiamondsBound int    `json:"diamondsBound"`
	Premium       bool   `json:"premium"`
	Vip           bool   `json:"vip"`
	Nivel         int    `json:"nivel"`
	Erro          string `json:"erro,omitempty"`
	Personagens   []struct {
		ID   int    `json:"id"`
		Nome string `json:"nome"`
	} `json:"personagens"`
}

type EscolherPersonagemResponse struct {
	Token string `json:"token"`
}

// EscolherPersonagem seleciona o personagem e retorna o token de sessão do personagem
func (c *PokeTrueClient) EscolherPersonagem(ctx context.Context, pjID int) (*EscolherPersonagemResponse, error) {
	payload := map[string]any{"personagem": pjID}
	var resp EscolherPersonagemResponse
	err := c.DoRequest(ctx, http.MethodPost, "personagens/escolher", payload, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Token != "" {
		c.token = resp.Token
	}
	return &resp, nil
}

// GetMe obtém os dados do perfil logado
func (c *PokeTrueClient) GetMe(ctx context.Context) (*MeResponse, error) {
	var resp MeResponse
	err := c.DoRequest(ctx, http.MethodGet, "me", nil, &resp)
	return &resp, err
}

// GetState obtém o estado completo e canônico do jogo
func (c *PokeTrueClient) GetState(ctx context.Context) (*PokeTrueState, error) {
	var st PokeTrueState
	err := c.DoRequest(ctx, http.MethodGet, "state", nil, &st)
	return &st, err
}

type HuntStartResponse struct {
	State *PokeTrueState `json:"state"`
	Erro  string         `json:"erro,omitempty"`
}

// StartHunt inicia a caçada numa determinada rota/hunt
func (c *PokeTrueClient) StartHunt(ctx context.Context, mapa string, auto map[string]any, offline bool) (*HuntStartResponse, error) {
	payload := map[string]any{
		"mapa":    mapa,
		"auto":    auto,
		"offline": offline,
	}
	var resp HuntStartResponse
	err := c.DoRequest(ctx, http.MethodPost, "hunt/start", payload, &resp)
	return &resp, err
}

type HuntSyncResponse struct {
	State *PokeTrueState `json:"state"`
	Delta *DeltaState    `json:"delta"`
	On    int            `json:"on"`
	Erro  string         `json:"erro,omitempty"`
}

// SyncHunt sincroniza o progresso da caçada
func (c *PokeTrueClient) SyncHunt(ctx context.Context, hp int, offline bool) (*HuntSyncResponse, error) {
	payload := map[string]any{
		"cena":         "farm",
		"hp":           hp,
		"offlineAtivo": offline,
		"enxuto":       1,
		"leve":         1,
	}
	var resp HuntSyncResponse
	err := c.DoRequest(ctx, http.MethodPost, "hunt/sync", payload, &resp)
	return &resp, err
}

// KillMonster registra a derrota de um monstro da fila
func (c *PokeTrueClient) KillMonster(ctx context.Context, filaID string, hpRatio float64) (*KillResponse, error) {
	payload := map[string]any{
		"filaId": filaID,
		"hp":     hpRatio,
		"enxuto": 1,
	}
	var resp KillResponse
	err := c.DoRequest(ctx, http.MethodPost, "hunt/kill", payload, &resp)
	return &resp, err
}

type ThrowResponse struct {
	State     *PokeTrueState `json:"state"`
	Capturado *CapturadoInfo `json:"capturado"`
	Erro      string         `json:"erro,omitempty"`
}

// ThrowBall arremessa uma pokébola num monstro da fila
func (c *PokeTrueClient) ThrowBall(ctx context.Context, bola string, filaID string) (*ThrowResponse, error) {
	payload := map[string]any{
		"bola":   bola,
		"filaId": filaID,
	}
	var resp ThrowResponse
	err := c.DoRequest(ctx, http.MethodPost, "hunt/throw", payload, &resp)
	return &resp, err
}

// StopHunt interrompe a caça
func (c *PokeTrueClient) StopHunt(ctx context.Context) error {
	return c.DoRequest(ctx, http.MethodPost, "hunt/stop", map[string]any{}, nil)
}

// SwitchActive define qual Pokémon do time fica ativo em campo
func (c *PokeTrueClient) SwitchActive(ctx context.Context, pokemonID int) error {
	payload := map[string]any{
		"pokemonId": pokemonID,
	}
	return c.DoRequest(ctx, http.MethodPost, "team/active", payload, nil)
}

// BuyItem compra itens no NPC Mark (loja)
func (c *PokeTrueClient) BuyItem(ctx context.Context, item string, qtd int) error {
	payload := map[string]any{
		"item": item,
		"qtd":  qtd,
	}
	return c.DoRequest(ctx, http.MethodPost, "shop/buy", payload, nil)
}

// SellLote vende uma lista de itens em lote no NPC Mark
func (c *PokeTrueClient) SellLote(ctx context.Context, lista []SellItem) error {
	payload := map[string]any{
		"lista": lista,
	}
	return c.DoRequest(ctx, http.MethodPost, "shop/sell-lote", payload, nil)
}

// ConsumeItem consome um item do stash (ex: candy)
func (c *PokeTrueClient) ConsumeItem(ctx context.Context, item string, qtd int) error {
	payload := map[string]any{
		"item": item,
		"qtd":  qtd,
	}
	return c.DoRequest(ctx, http.MethodPost, "item/consumir", payload, nil)
}

// GetHelds obtém todos os held items equipados e na mochila
func (c *PokeTrueClient) GetHelds(ctx context.Context) (*HeldsResponse, error) {
	var resp HeldsResponse
	err := c.DoRequest(ctx, http.MethodGet, "helds", nil, &resp)
	return &resp, err
}

// EquipHeld equipa um held item num Pokémon
func (c *PokeTrueClient) EquipHeld(ctx context.Context, pokemonID, heldID int) error {
	payload := map[string]any{
		"pokemonId": pokemonID,
		"heldId":    heldID,
	}
	return c.DoRequest(ctx, http.MethodPost, "helds/equipar", payload, nil)
}

// FuseHelds realiza fusão de held items do mesmo tier
func (c *PokeTrueClient) FuseHelds(ctx context.Context, ids []int) error {
	payload := map[string]any{
		"ids": ids,
	}
	return c.DoRequest(ctx, http.MethodPost, "helds/fundir", payload, nil)
}

// GetTreino consulta o campo de treinamento de stamina/EVs
func (c *PokeTrueClient) GetTreino(ctx context.Context, lugar string) (*TreinoResponse, error) {
	var resp TreinoResponse
	err := c.DoRequest(ctx, http.MethodGet, "treino?lugar="+lugar, nil, &resp)
	return &resp, err
}

// SetTreino coloca um Pokémon de reserva para treinar atributo
func (c *PokeTrueClient) SetTreino(ctx context.Context, lugar string, slot, pokemonID int, stat string) error {
	payload := map[string]any{
		"lugar":     lugar,
		"slot":      slot,
		"pokemonId": pokemonID,
		"stat":      stat,
	}
	return c.DoRequest(ctx, http.MethodPost, "treino/por", payload, nil)
}

// RefillTreino recarrega as cargas de treino com ouro
func (c *PokeTrueClient) RefillTreino(ctx context.Context, lugar string, slot int, tipo string) error {
	payload := map[string]any{
		"lugar": lugar,
		"slot":  slot,
		"tipo":  tipo,
	}
	return c.DoRequest(ctx, http.MethodPost, "treino/recarregar", payload, nil)
}

// GetJenny obtém as missões da Oficial Jenny
func (c *PokeTrueClient) GetJenny(ctx context.Context) (*JennyResponse, error) {
	var resp JennyResponse
	err := c.DoRequest(ctx, http.MethodGet, "jenny", nil, &resp)
	return &resp, err
}

// ChooseJenny escolhe as missões diárias da Jenny
func (c *PokeTrueClient) ChooseJenny(ctx context.Context, especies []string) error {
	payload := map[string]any{
		"especies": especies,
	}
	return c.DoRequest(ctx, http.MethodPost, "jenny/escolher", payload, nil)
}

// GetTarefas lista as tarefas gerais disponíveis
func (c *PokeTrueClient) GetTarefas(ctx context.Context) (*TarefasResponse, error) {
	var resp TarefasResponse
	err := c.DoRequest(ctx, http.MethodGet, "tarefas", nil, &resp)
	return &resp, err
}

// CompleteTarefa conclui uma tarefa e recolhe os prêmios
func (c *PokeTrueClient) CompleteTarefa(ctx context.Context, id string) error {
	payload := map[string]any{
		"id": id,
	}
	return c.DoRequest(ctx, http.MethodPost, "tarefas/concluir", payload, nil)
}

// GetGrupo obtém o status da party/grupo
func (c *PokeTrueClient) GetGrupo(ctx context.Context) (*GroupState, error) {
	var resp GroupState
	err := c.DoRequest(ctx, http.MethodGet, "grupo", nil, &resp)
	return &resp, err
}

// CreateGrupo cria uma nova party definindo o Pokémon e papel (tanque, suporte, dps)
func (c *PokeTrueClient) CreateGrupo(ctx context.Context, pokemonID int, papel string) (*GroupState, error) {
	payload := map[string]any{
		"pokemonId": pokemonID,
		"papel":     papel,
	}
	var resp GroupState
	err := c.DoRequest(ctx, http.MethodPost, "grupo/criar", payload, &resp)
	return &resp, err
}

// InviteGrupo convida outro jogador/bot para a party pelo nome do personagem
func (c *PokeTrueClient) InviteGrupo(ctx context.Context, nome string) (*GroupState, error) {
	payload := map[string]any{
		"nome": nome,
	}
	var resp GroupState
	err := c.DoRequest(ctx, http.MethodPost, "grupo/convidar", payload, &resp)
	return &resp, err
}

// RespondGrupo aceita ou recusa um convite de party
func (c *PokeTrueClient) RespondGrupo(ctx context.Context, grupoID int, aceitar bool, pokemonID int, papel string) (*GroupState, error) {
	payload := map[string]any{
		"grupo":   grupoID,
		"aceitar": aceitar,
	}
	if aceitar {
		payload["pokemonId"] = pokemonID
		payload["papel"] = papel
	}
	var resp GroupState
	err := c.DoRequest(ctx, http.MethodPost, "grupo/responder", payload, &resp)
	return &resp, err
}

// StartGrupo inicia a caçada da party (apenas líder)
func (c *PokeTrueClient) StartGrupo(ctx context.Context) (*GroupState, error) {
	var resp GroupState
	err := c.DoRequest(ctx, http.MethodPost, "grupo/comecar", map[string]any{}, &resp)
	return &resp, err
}

// LeaveGrupo sai da party
func (c *PokeTrueClient) LeaveGrupo(ctx context.Context) error {
	return c.DoRequest(ctx, http.MethodPost, "grupo/sair", map[string]any{}, nil)
}

// StopGrupo encerra a caçada em grupo
func (c *PokeTrueClient) StopGrupo(ctx context.Context) error {
	return c.DoRequest(ctx, http.MethodPost, "grupo/parar", map[string]any{}, nil)
}

// ResgatarPokeLog recolhe recompensas de PokéLog (Bestiário) concluídas
func (c *PokeTrueClient) ResgatarPokeLog(ctx context.Context, tudo bool, especie string) (*PokeLogResgatarResponse, error) {
	payload := map[string]any{}
	if tudo {
		payload["tudo"] = true
	} else if especie != "" {
		payload["especie"] = especie
	}
	var resp PokeLogResgatarResponse
	err := c.DoRequest(ctx, http.MethodPost, "pokelog/resgatar", payload, &resp)
	if err != nil {
		if strings.Contains(err.Error(), "nada_a_resgatar") {
			return &resp, nil
		}
		return nil, err
	}
	return &resp, nil
}

// GetRunas consulta as runas, pontos e slots do treinador
func (c *PokeTrueClient) GetRunas(ctx context.Context) (*RunasResponse, error) {
	var resp RunasResponse
	err := c.DoRequest(ctx, http.MethodGet, "runas", nil, &resp)
	return &resp, err
}

// ComprarRuna adquire uma runa na loja de runas com pontos de PokéLog
func (c *PokeTrueClient) ComprarRuna(ctx context.Context, runa string) error {
	payload := map[string]any{"runa": runa}
	return c.DoRequest(ctx, http.MethodPost, "runas/comprar", payload, nil)
}

// EncaixarRuna equipa uma runa num slot ativo associando a um elemento
func (c *PokeTrueClient) EncaixarRuna(ctx context.Context, slot int, runa, elemento string) error {
	payload := map[string]any{
		"slot":     slot,
		"runa":     runa,
		"elemento": elemento,
	}
	return c.DoRequest(ctx, http.MethodPost, "runas/encaixar", payload, nil)
}

// EvolvePokemon realiza a evolução de um Pokémon
func (c *PokeTrueClient) EvolvePokemon(ctx context.Context, pokemonID, indice int) (*EvolveResponse, error) {
	payload := map[string]any{
		"pokemonId": pokemonID,
		"indice":    indice,
	}
	var resp EvolveResponse
	err := c.DoRequest(ctx, http.MethodPost, "pokemon/evolve", payload, &resp)
	return &resp, err
}

// GetGinasios consulta os ginásios disponíveis e status das insígnias
func (c *PokeTrueClient) GetGinasios(ctx context.Context) (*GinasiosResponse, error) {
	var resp GinasiosResponse
	err := c.DoRequest(ctx, http.MethodGet, "ginasio", nil, &resp)
	return &resp, err
}

// DesafiarGinasio desafia o líder de um ginásio
func (c *PokeTrueClient) DesafiarGinasio(ctx context.Context, tipo string) (*DesafiarGinasioResponse, error) {
	payload := map[string]any{
		"tipo":   tipo,
		"contra": "lider",
	}
	var resp DesafiarGinasioResponse
	err := c.DoRequest(ctx, http.MethodPost, "ginasio/desafiar", payload, &resp)
	return &resp, err
}

// GetBossInfo consulta o status de um World Boss diário
func (c *PokeTrueClient) GetBossInfo(ctx context.Context, chefe string) (*BossInfoResponse, error) {
	var resp BossInfoResponse
	err := c.DoRequest(ctx, http.MethodGet, "boss/info?chefe="+chefe, nil, &resp)
	return &resp, err
}

// LutarBoss desafia um World Boss diário usando a tentativa grátis
func (c *PokeTrueClient) LutarBoss(ctx context.Context, chefe string, pokemonID int) (*BossLutaResponse, error) {
	payload := map[string]any{
		"chefe":       chefe,
		"dificuldade": "facil",
		"pokemonId":   pokemonID,
		"pagar":       false,
		"leite":       false,
	}
	var resp BossLutaResponse
	err := c.DoRequest(ctx, http.MethodPost, "boss/lutar", payload, &resp)
	return &resp, err
}

