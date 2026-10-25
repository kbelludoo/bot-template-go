package engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// StateOverlay gerencia o estado bruto do jogo preservando 100% do schema JSON do servidor
// Evita o erro de "save malformado" ou "recusado: true" ao enviar saves parciais
type StateOverlay struct {
	mu            sync.RWMutex
	raw           map[string]any
	rev           int64
	saveVersion   int64
	clientVersion string
	ownedFields   []string
	counters      map[string]int64
}

// NewStateOverlay inicializa o gerenciador a partir do mapa bruto retornado pelo servidor
func NewStateOverlay(raw map[string]any) *StateOverlay {
	if raw == nil {
		raw = make(map[string]any)
	}

	so := &StateOverlay{
		raw:      deepCopyMap(raw),
		counters: make(map[string]int64),
	}

	// Tentar capturar rev ou saveVersion inicial se existir
	if r, ok := so.GetInt64("rev"); ok {
		so.rev = r
	} else if r, ok := so.GetInt64("revisao"); ok {
		so.rev = r
	}
	if sv, ok := so.GetInt64("saveVersion"); ok {
		so.saveVersion = sv
	}

	return so
}

// LoadRaw atualiza o estado bruto a partir de um JSON ou resposta de GET oficial
func (s *StateOverlay) LoadRaw(rawJSON []byte) error {
	var m map[string]any
	if err := json.Unmarshal(rawJSON, &m); err != nil {
		return fmt.Errorf("falha ao decodificar JSON do estado canônico: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if sv, ok := getInt64FromMap(m, "saveVersion"); ok {
		s.saveVersion = sv
	}
	if cv, ok := m["clientVersion"].(string); ok && cv != "" {
		s.clientVersion = cv
	}
	if omitted, ok := m["saveOwnedFields"].([]any); ok {
		fields := make([]string, 0, len(omitted))
		for _, item := range omitted {
			if str, ok := item.(string); ok {
				fields = append(fields, str)
			}
		}
		s.ownedFields = fields
	}

	// Se o JSON veio envelopado em {"estado": ...} ou {"save": ...}, desempacotar
	if est, ok := m["estado"].(map[string]any); ok {
		s.raw = deepCopyMap(est)
	} else if sav, ok := m["save"].(map[string]any); ok {
		s.raw = deepCopyMap(sav)
	} else {
		s.raw = deepCopyMap(m)
	}

	if r, ok := getInt64FromMap(s.raw, "rev"); ok {
		s.rev = r
	}
	return nil
}

// Set atualiza um valor por chave simples ou caminho separado por pontos (ex: "totais.abates" ou "ouro")
func (s *StateOverlay) Set(path string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		s.raw[parts[0]] = val
		return
	}

	curr := s.raw
	for i := 0; i < len(parts)-1; i++ {
		p := parts[i]
		nested, ok := curr[p].(map[string]any)
		if !ok {
			nested = make(map[string]any)
			curr[p] = nested
		}
		curr = nested
	}
	curr[parts[len(parts)-1]] = val
}

// Get recupera um valor do mapa bruto
func (s *StateOverlay) Get(path string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parts := strings.Split(path, ".")
	curr := any(s.raw)

	for _, p := range parts {
		m, ok := curr.(map[string]any)
		if !ok {
			return nil, false
		}
		val, exists := m[p]
		if !exists {
			return nil, false
		}
		curr = val
	}
	return curr, true
}

// GetInt64 helper para ler inteiros de forma segura independente do tipo float64 do JSON
func (s *StateOverlay) GetInt64(path string) (int64, bool) {
	val, ok := s.Get(path)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	default:
		return 0, false
	}
}

// Merge mescla um mapa de alterações sobre o mapa bruto original sem apagar chaves desconhecidas
func (s *StateOverlay) Merge(diff map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	deepMergeMaps(s.raw, diff)
}

// Envelope constrói o envelope para salvar via PUT /api/save (ex: {"estado": raw})
func (s *StateOverlay) Envelope(key string) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copyRaw := deepCopyMap(s.raw)
	if key == "" {
		return copyRaw
	}
	return map[string]any{
		key: copyRaw,
	}
}

// ContractEnvelope constrói o envelope completo com controle de versão e campos omitidos (padrão PokeLegacy)
func (s *StateOverlay) ContractEnvelope(stateKey string) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if stateKey == "" {
		stateKey = "save"
	}
	env := map[string]any{
		stateKey: deepCopyMap(s.raw),
	}
	if s.saveVersion > 0 {
		env["saveVersion"] = s.saveVersion
	}
	if s.clientVersion != "" {
		env["clientVersion"] = s.clientVersion
	}
	if len(s.ownedFields) > 0 {
		omitted := make([]string, len(s.ownedFields))
		copy(omitted, s.ownedFields)
		env["omitted"] = omitted
	}
	return env
}

// ToJSON serializa o envelope completo para envio HTTP
func (s *StateOverlay) ToJSON(envelopeKey string) ([]byte, error) {
	env := s.Envelope(envelopeKey)
	return json.Marshal(env)
}

// UpdateRev atualiza a revisão do documento para controle de concorrência
func (s *StateOverlay) UpdateRev(newRev int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rev = newRev
	s.raw["rev"] = newRev
}

// Rev retorna a revisão atual
func (s *StateOverlay) Rev() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rev
}

// SetSaveContract define os metadados de versão de contrato (usado em sistemas SAVE_CONTRACT_STALE)
func (s *StateOverlay) SetSaveContract(saveVersion int64, clientVersion string, ownedFields []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saveVersion = saveVersion
	if clientVersion != "" {
		s.clientVersion = clientVersion
	}
	if ownedFields != nil {
		s.ownedFields = append([]string(nil), ownedFields...)
	}
}

// SaveVersion retorna a versão atual do contrato de save
func (s *StateOverlay) SaveVersion() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveVersion
}

// IsStaleContractConflict verifica se a resposta HTTP 409 representa apenas descompasso de versão (SAVE_CONTRACT_STALE)
// que deve ser reconciliado com GET /api/me ou /api/eu em vez de desarmar o Circuit Breaker
func IsStaleContractConflict(statusCode int, body string) bool {
	if statusCode != 409 {
		return false
	}
	lower := strings.ToLower(body)
	return strings.Contains(lower, "save_contract_stale") ||
		strings.Contains(lower, "progresso desatualizado") ||
		strings.Contains(lower, "version mismatch") ||
		strings.Contains(lower, "rev")
}

// ReconcileStaleSave recarrega o estado oficial do servidor após HTTP 409 e reaplica as mutações locais com segurança
func (s *StateOverlay) ReconcileStaleSave(fetchCanonical func() ([]byte, error), reapplyLocal func(*StateOverlay)) error {
	rawBytes, err := fetchCanonical()
	if err != nil {
		return fmt.Errorf("falha ao buscar estado oficial para reconciliação 409: %w", err)
	}
	if err := s.LoadRaw(rawBytes); err != nil {
		return err
	}
	if reapplyLocal != nil {
		reapplyLocal(s)
	}
	return nil
}

// CheckMonotonicity valida a regra de ouro anti-cheat: contadores nunca devem regredir (Δ >= 0)
// Protege contra suspensão por "contadores voltaram para trás" ou "relógio inválido"
func (s *StateOverlay) CheckMonotonicity(counterKey string, newVal int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastVal, exists := s.counters[counterKey]
	if !exists {
		// Se não havia registro em memória, buscar do raw
		if rVal, ok := getInt64FromMap(s.raw, counterKey); ok {
			lastVal = rVal
		}
	}

	if newVal < lastVal {
		return fmt.Errorf("violação de monotonicidade em '%s': novo valor (%d) < valor anterior (%d)", counterKey, newVal, lastVal)
	}

	s.counters[counterKey] = newVal
	return nil
}

// VerifyAttributeBudget valida se os pontos de atributo alocados respeitam o teto da fórmula do jogo
// Σ atrib + pontosAtrib <= PONTOS_POR_NIVEL * (nível - 1)
func VerifyAttributeBudget(spentAttributes, freePoints, level, pointsPerLevel int) error {
	if level <= 1 {
		if spentAttributes > 0 {
			return fmt.Errorf("nível 1 não pode ter atributos distribuídos")
		}
		return nil
	}
	maxBudget := pointsPerLevel * (level - 1)
	total := spentAttributes + freePoints
	if total > maxBudget {
		return fmt.Errorf("orçamento de atributos excedido (%d > %d)", total, maxBudget)
	}
	return nil
}

// deepCopyMap cria uma cópia profunda de um mapa
func deepCopyMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		if nested, ok := v.(map[string]any); ok {
			dst[k] = deepCopyMap(nested)
		} else if slice, ok := v.([]any); ok {
			dst[k] = deepCopySlice(slice)
		} else {
			dst[k] = v
		}
	}
	return dst
}

func deepCopySlice(src []any) []any {
	dst := make([]any, len(src))
	for i, v := range src {
		if nestedMap, ok := v.(map[string]any); ok {
			dst[i] = deepCopyMap(nestedMap)
		} else if nestedSlice, ok := v.([]any); ok {
			dst[i] = deepCopySlice(nestedSlice)
		} else {
			dst[i] = v
		}
	}
	return dst
}

func deepMergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstMap, ok := dst[k].(map[string]any); ok {
				deepMergeMaps(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}

func getInt64FromMap(m map[string]any, path string) (int64, bool) {
	parts := strings.Split(path, ".")
	var curr any = m
	for _, p := range parts {
		nested, ok := curr.(map[string]any)
		if !ok {
			return 0, false
		}
		val, exists := nested[p]
		if !exists {
			return 0, false
		}
		curr = val
	}
	switch v := curr.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}
