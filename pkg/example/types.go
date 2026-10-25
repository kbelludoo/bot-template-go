package example

// AccountConfig modelo de configuração lido do accounts.json
type AccountConfig struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Enabled     bool            `json:"enabled"`
	Credentials GameCredentials `json:"credentials"`
	Settings    GameSettings    `json:"settings"`
	BindIP      string          `json:"bindIP,omitempty"`
}

type GameCredentials struct {
	Token         string `json:"token"`
	SessionCookie string `json:"sessionCookie,omitempty"`
	Email         string `json:"email,omitempty"`
	Password      string `json:"password,omitempty"`
}

type GameSettings struct {
	HuntTarget     string  `json:"huntTarget"` // "auto" ou slug da área
	AutoHunt       bool    `json:"autoHunt"`
	AutoSell       bool    `json:"autoSell"`
	AutoHeal       bool    `json:"autoHeal"`
	AutoDaily      bool    `json:"autoDaily"`
	AutoPromotion  bool    `json:"autoPromotion"`
	AutoDummies    bool    `json:"autoDummies"`
	HealThreshold  float64 `json:"healThreshold"` // ex: 0.5 (cura quando HP < 50%)
	MinGoldReserve int64   `json:"minGoldReserve"`
}

// GameState modelo canônico de dados do jogo
type GameState struct {
	Level       int            `json:"level"`
	Exp         int64          `json:"exp"`
	Gold        int64          `json:"gold"`
	HP          int            `json:"hp"`
	MaxHP       int            `json:"maxHp"`
	Mana        int            `json:"mana"`
	MaxMana     int            `json:"maxMana"`
	Vocation    string         `json:"vocation"`
	CurrentArea string         `json:"currentArea"`
	Inventory   map[string]int `json:"inventory"`
}
