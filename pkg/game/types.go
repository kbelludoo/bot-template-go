package game

// AccountConfig modelo de configuração do bot lido do accounts.json
type AccountConfig struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Enabled     bool            `json:"enabled"`
	Credentials GameCredentials `json:"credentials"`
	Settings    GameSettings    `json:"settings"`
	BindIP      string          `json:"bindIP,omitempty"`
}

type GameCredentials struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	SessionCookie string `json:"sessionCookie,omitempty"`
	Token         string `json:"token"`
}

type GameSettings struct {
	HuntTarget    string  `json:"huntTarget"` // "auto" para resolver automaticamente a melhor área ou slug específico
	AutoHunt      bool    `json:"autoHunt"`
	AutoSell      bool    `json:"autoSell"`
	AutoHeal      bool    `json:"autoHeal"`
	AutoDaily     bool    `json:"autoDaily"`
	AutoEquip     bool    `json:"autoEquip"`
	AutoTreino    bool    `json:"autoTreino"`
	AutoCatch     bool    `json:"autoCatch"`
	BallType      string  `json:"ballType"`      // "auto", "pokeball", "greatball", "superball", "ultraball"
	HealThreshold float64 `json:"healThreshold"` // ex: 0.5 (cura quando HP < 50%)
	MinGoldKeep   int64   `json:"minGoldKeep"`   // reserva de ouro mantida
}

// PokeTrueState expõe os campos conhecidos e preserva o JSON completo em RawEstado.
type PokeTrueState struct {
	RawEstado     map[string]any      `json:"-"`
	Gold          int64               `json:"gold"`
	Nivel         int                 `json:"nivel"`
	NivelXp       int                 `json:"nivelXp"`
	Pokemon       []PokemonState      `json:"pokemon"`
	Hunt          HuntState           `json:"hunt"`
	Itens         map[string]int      `json:"itens"`
	Vip           bool                `json:"vip"`
	GM            bool                `json:"gm"`
	Cla           *ClanState          `json:"cla,omitempty"`
	LiderGin      *LiderState         `json:"liderGin,omitempty"`
	InsigniasGin  *InsigniasState     `json:"insigniasGin,omitempty"`
	PokelogPontos int                 `json:"pokelogPontos"`
	Dex           map[string]DexEntry `json:"dex,omitempty"`
	Erro          string              `json:"erro,omitempty"`
}

type PokemonState struct {
	ID       int            `json:"id"`
	Especie  string         `json:"especie"`
	Nivel    int            `json:"nivel"`
	Xp       int            `json:"xp"`
	Hp       float64        `json:"hp"`
	HpMax    int            `json:"hpMax,omitempty"`
	TeamSlot *int           `json:"teamSlot"`
	Shiny    bool           `json:"shiny"`
	Held     *HeldState     `json:"held"`
	IVs      map[string]int `json:"ivs,omitempty"`
	EVs      map[string]int `json:"evs,omitempty"`
	Boost    int            `json:"boost,omitempty"`
	Natureza string         `json:"natureza,omitempty"`
}

type HeldState struct {
	ID    int    `json:"id,omitempty"`
	Tipo  string `json:"tipo"`
	Tier  int    `json:"tier"`
	Nome  string `json:"nome,omitempty"`
	Texto string `json:"texto,omitempty"`
}

type HuntState struct {
	Mapa    string         `json:"mapa"`
	AtivoID int            `json:"ativoId"`
	Fila    []FilaMonster  `json:"fila"`
	Auto    map[string]any `json:"auto,omitempty"`
}

type FilaMonster struct {
	ID        string `json:"id"`
	Especie   string `json:"especie"`
	Nivel     int    `json:"nivel"`
	Shiny     bool   `json:"shiny"`
	Morto     bool   `json:"morto"`
	Esperando bool   `json:"esperando"`
	RestaMs   *int   `json:"restaMs"`
}

type ClanState struct {
	Slug  string   `json:"slug"`
	Rank  int      `json:"rank"`
	Tipos []string `json:"tipos"`
	Pct   float64  `json:"pct"`
}

type LiderState struct {
	Tipos []string `json:"tipos"`
	Pct   float64  `json:"pct"`
}

type InsigniasState struct {
	PorTipo map[string]float64 `json:"porTipo"`
}

// KillResponse retornado por /api/hunt/kill
type KillResponse struct {
	Delta      *DeltaState    `json:"delta"`
	State      *PokeTrueState `json:"state,omitempty"`
	Capturado  *CapturadoInfo `json:"capturado,omitempty"`
	Substituto *PokemonState  `json:"substituto,omitempty"`
	Erro       string         `json:"erro,omitempty"`
}

type DeltaState struct {
	Kills     int             `json:"kills"`
	Xp        int             `json:"xp"`
	XpPoke    int             `json:"xpPoke"`
	Gold      int64           `json:"gold"`
	Loot      map[string]int  `json:"loot"`
	Capturas  []CapturadoInfo `json:"capturas,omitempty"`
	Treinador *TreinadorDelta `json:"treinador,omitempty"`
	Poke      *PokeDelta      `json:"poke,omitempty"`
	ExpShare  *ExpShareDelta  `json:"expShare,omitempty"`
}

type TreinadorDelta struct {
	Nivel int `json:"nivel"`
	Xp    int `json:"xp"`
}

type PokeDelta struct {
	ID    int `json:"id"`
	Nivel int `json:"nivel"`
	Xp    int `json:"xp"`
}

type ExpShareDelta struct {
	Todos []PokeDelta `json:"todos"`
}

type CapturadoInfo struct {
	Especie string `json:"especie"`
	Nivel   int    `json:"nivel"`
	Shiny   bool   `json:"shiny"`
}

// HeldsResponse retornado por /api/helds
type HeldsResponse struct {
	Mochila   []HeldDetail `json:"mochila"`
	Equipados []HeldDetail `json:"equipados"`
	Tokens    int          `json:"devotedTokens,omitempty"`
	Erro      string       `json:"erro,omitempty"`
}

type HeldDetail struct {
	ID        int    `json:"id"`
	Tipo      string `json:"tipo"`
	Tier      int    `json:"tier"`
	PokemonID *int   `json:"pokemonId"`
}

// TreinoResponse retornado por /api/treino
type TreinoResponse struct {
	Slots []TreinoSlot `json:"slots"`
	Erro  string       `json:"erro,omitempty"`
}

type TreinoSlot struct {
	Slot      int    `json:"slot"`
	PokemonID *int   `json:"pokemonId"`
	Stat      string `json:"stat"`
	Cargas    int    `json:"cargas"`
	MaxCargas int    `json:"maxCargas"`
	GastoOuro int64  `json:"gastoOuro,omitempty"`
}

// JennyResponse retornado por /api/jenny
type JennyResponse struct {
	Escolhidas bool          `json:"escolhidas"`
	Quantas    int           `json:"quantas"`
	Tarefas    []JennyTarefa `json:"tarefas"`
	Erro       string        `json:"erro,omitempty"`
}

type JennyTarefa struct {
	Especie  string `json:"especie"`
	Hunt     string `json:"hunt"`
	HuntNome string `json:"huntNome"`
	Meta     int    `json:"meta"`
	Feito    int    `json:"feito"`
	Pago     bool   `json:"pago"`
	Xp       int    `json:"xp"`
	Tokens   int    `json:"tokens"`
}

// TarefasResponse retornado por /api/tarefas
type TarefasResponse struct {
	Tarefas []TarefaItem `json:"tarefas"`
	Erro    string       `json:"erro,omitempty"`
}

type TarefaItem struct {
	ID        string `json:"id"`
	Titulo    string `json:"titulo"`
	Pronto    bool   `json:"pronto"`
	Concluido bool   `json:"concluido"`
}

// GroupState retornado por /api/grupo
type GroupState struct {
	Grupo    *GroupDetail  `json:"grupo"`
	Convites []GroupInvite `json:"convites"`
	Meus     []GroupMember `json:"meus"`
	Regras   *GroupRules   `json:"regras"`
	Erro     string        `json:"erro,omitempty"`
}

type GroupDetail struct {
	Lider      int           `json:"lider"`
	SouLider   bool          `json:"souLider"`
	Membros    []GroupMember `json:"membros"`
	Convidados []string      `json:"convidados"`
	Luta       *GroupFight   `json:"luta"`
}

type GroupMember struct {
	Pj        int    `json:"pj"`
	Nome      string `json:"nome"`
	PokemonID int    `json:"pokemonId"`
	Especie   string `json:"especie"`
	Nivel     int    `json:"nivel"`
	Shiny     bool   `json:"shiny"`
	Papel     string `json:"papel"` // "tanque", "suporte", "dps"
	Eu        bool   `json:"eu"`
}

type GroupInvite struct {
	Grupo   int           `json:"grupo"`
	De      string        `json:"de"`
	Membros []GroupMember `json:"membros"`
	VenceEm int64         `json:"venceEm"`
}

type GroupFight struct {
	Ativa bool `json:"ativa"`
}

type GroupRules struct {
	Nome     string   `json:"nome"`
	NivelMin int      `json:"nivelMin"`
	Especies []string `json:"especies"`
	VidaX    float64  `json:"vidaX"`
}

// DexEntry representa os dados de registro e abates de uma espécie no estado
type DexEntry struct {
	Visto     any `json:"v"`
	Capturado any `json:"c"`
	Abates    int `json:"a"`
	Coletado  int `json:"k"`
}

// PokeLogResgatarResponse retornado por /api/pokelog/resgatar
type PokeLogResgatarResponse struct {
	Feitos []PokeLogFeito `json:"feitos"`
	Xp     int            `json:"xp"`
	Pontos int            `json:"pontos"`
	State  *PokeTrueState `json:"state,omitempty"`
	Erro   string         `json:"erro,omitempty"`
}

type PokeLogFeito struct {
	Especie string `json:"especie"`
	Estagio int    `json:"estagio"`
	Xp      int    `json:"xp"`
	Pontos  int    `json:"pontos"`
}

// RunasResponse retornado por /api/runas
type RunasResponse struct {
	Pontos    int        `json:"pontos"`
	Compradas []string   `json:"compradas"`
	Slots     []RunaSlot `json:"slots"`
	Erro      string     `json:"erro,omitempty"`
}

type RunaSlot struct {
	Runa     string `json:"runa"`
	Elemento string `json:"elemento"`
}

// EvolveResponse retornado por /api/pokemon/evolve
type EvolveResponse struct {
	State *PokeTrueState `json:"state,omitempty"`
	Erro  string         `json:"erro,omitempty"`
}

// GinasiosResponse retornado por /api/ginasio
type GinasiosResponse struct {
	MeuNivel int           `json:"meuNivel"`
	NivelMin int           `json:"nivelMin"`
	Lista    []GinasioItem `json:"lista"`
	Erro     string        `json:"erro,omitempty"`
}

type GinasioItem struct {
	Tipo        string `json:"tipo"`
	Nome        string `json:"nome"`
	NivelMin    int    `json:"nivelMin"`
	TemInsignia bool   `json:"temInsignia"`
	Tentativas  struct {
		GratisRestam int `json:"gratisRestam"`
		TotalRestam  int `json:"totalRestam"`
	} `json:"tentativas"`
}

// DesafiarGinasioResponse retornado por /api/ginasio/desafiar
type DesafiarGinasioResponse struct {
	Venceu bool           `json:"venceu"`
	State  *PokeTrueState `json:"state,omitempty"`
	Erro   string         `json:"erro,omitempty"`
}

// BossInfoResponse retornado por /api/boss/info
type BossInfoResponse struct {
	Entradas struct {
		GratisRestam int `json:"gratisRestam"`
	} `json:"entradas"`
	Erro string `json:"erro,omitempty"`
}

// BossLutaResponse retornado por /api/boss/lutar
type BossLutaResponse struct {
	Venceu bool           `json:"venceu"`
	Loot   map[string]int `json:"loot"`
	Xp     int            `json:"xp"`
	Gold   int64          `json:"gold"`
	Erro   string         `json:"erro,omitempty"`
}

// VipBuyResponse retornado por /api/vip/buy
type VipBuyResponse struct {
	Saldo  int    `json:"saldo"`
	VipAte int64  `json:"vipAte"`
	Erro   string `json:"erro,omitempty"`
}

// MarketListingsResponse retornado por /api/market
type MarketListingsResponse struct {
	Anuncios []MarketListingItem `json:"anuncios"`
	Total    int                 `json:"total"`
	Erro     string              `json:"erro,omitempty"`
}

type MarketListingItem struct {
	ID       int    `json:"id"`
	Tipo     string `json:"tipo"`
	Moeda    string `json:"moeda"`
	Preco    int64  `json:"preco"`
	ItemQtd  int    `json:"itemQtd"`
	Avulso   bool   `json:"avulso"`
	Vendedor string `json:"vendedor"`
	Modo     string `json:"modo"`
}

// MarketBuyResponse retornado por /api/market/buy
type MarketBuyResponse struct {
	Tipo    string         `json:"tipo"`
	ItemQtd int            `json:"itemQtd"`
	Preco   int64          `json:"preco"`
	State   *PokeTrueState `json:"state,omitempty"`
	Erro    string         `json:"erro,omitempty"`
}

