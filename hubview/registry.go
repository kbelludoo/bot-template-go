package hubview

import (
	"fmt"
	"strings"
)

// BotMeta defines visual presentation properties for a bot
type BotMeta struct {
	Key          string
	Icon         string
	Name         string
	Game         string
	Port         int
	Theme        string
	ThemeBg      string
	ThemeBorder  string
	ThemeGlow    string
	DashboardURL string
	Order        int
	IsStream     bool
	IsService    bool
}

var knownMetas = map[string]BotMeta{
	"huntera": {
		Key: "huntera", Icon: "⚔️", Name: "Huntera Idle", Game: "Huntera Idle • MMORPG",
		Port: 3000, Theme: "#ef4444", ThemeBg: "rgba(239, 68, 68, 0.12)",
		ThemeBorder: "rgba(239, 68, 68, 0.45)", ThemeGlow: "rgba(239, 68, 68, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/huntera-bot/", Order: 1,
	},
	"tibidle": {
		Key: "tibidle", Icon: "🛡️", Name: "Tibidle Multi-Bot", Game: "Tibidle • Canvas Idle RPG",
		Port: 3001, Theme: "#00f0ff", ThemeBg: "rgba(0, 240, 255, 0.12)",
		ThemeBorder: "rgba(0, 240, 255, 0.45)", ThemeGlow: "rgba(0, 240, 255, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/tibidle-bot/", Order: 2,
	},
	"sigma": {
		Key: "sigma", Icon: "⚡", Name: "Pokemon Sigma Multi-Bot", Game: "Pokemon Sigma • Web RPG",
		Port: 3002, Theme: "#facc15", ThemeBg: "rgba(250, 204, 21, 0.12)",
		ThemeBorder: "rgba(250, 204, 21, 0.45)", ThemeGlow: "rgba(250, 204, 21, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/sigma-bot/", Order: 3,
	},
	"leveling": {
		Key: "leveling", Icon: "🌲", Name: "Leveling Idle 2D", Game: "Leveling Idle 2D • Forest MMORPG",
		Port: 3003, Theme: "#10b981", ThemeBg: "rgba(16, 185, 129, 0.12)",
		ThemeBorder: "rgba(16, 185, 129, 0.45)", ThemeGlow: "rgba(16, 185, 129, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/leveling-bot/", Order: 4,
	},
	"pokelegacy": {
		Key: "pokelegacy", Icon: "⚡", Name: "PokeLegacy Online", Game: "PokeLegacy Online • Web RPG (5x)",
		Port: 3005, Theme: "#fbbf24", ThemeBg: "rgba(251, 191, 36, 0.12)",
		ThemeBorder: "rgba(251, 191, 36, 0.45)", ThemeGlow: "rgba(251, 191, 36, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/pokelegacy-bot/", Order: 5,
	},
	"pokeidle": {
		Key: "pokeidle", Icon: "🎮", Name: "PokeIdle.io Multi-Bot", Game: "PokeIdle.io • Classic Canvas",
		Port: 3006, Theme: "#38bdf8", ThemeBg: "rgba(56, 189, 248, 0.12)",
		ThemeBorder: "rgba(56, 189, 248, 0.45)", ThemeGlow: "rgba(56, 189, 248, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/pokeidle-bot/", Order: 6,
	},
	"pokeidleworld": {
		Key: "pokeidleworld", Icon: "🌍", Name: "Poke Idle World", Game: "Poke Idle World • MMO Exploration",
		Port: 3007, Theme: "#34d399", ThemeBg: "rgba(52, 211, 153, 0.12)",
		ThemeBorder: "rgba(52, 211, 153, 0.45)", ThemeGlow: "rgba(52, 211, 153, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/pokeidleworld-bot/", Order: 7,
	},
	"pokepixel": {
		Key: "pokepixel", Icon: "👾", Name: "PokePixel Multi-Bot", Game: "PokePixel Online • Pixel Canvas RPG",
		Port: 3015, Theme: "#ec4899", ThemeBg: "rgba(236, 72, 153, 0.12)",
		ThemeBorder: "rgba(236, 72, 153, 0.45)", ThemeGlow: "rgba(236, 72, 153, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/pokepixel-bot/", Order: 8,
	},
	"poketrue": {
		Key: "poketrue", Icon: "🔥", Name: "PokeTrue Multi-Bot", Game: "PokeTrue Online • Weakness Solver & Idle Hunt",
		Port: 3014, Theme: "#ef4444", ThemeBg: "rgba(239, 68, 68, 0.12)",
		ThemeBorder: "rgba(239, 68, 68, 0.45)", ThemeGlow: "rgba(239, 68, 68, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/bot-template-go/?v=force1", Order: 9,
	},
	"baiak": {
		Key: "baiak", Icon: "🔮", Name: "Baiak Idle Multi-Bot", Game: "Baiak Idle • Open Tibia Idle",
		Port: 8090, Theme: "#c084fc", ThemeBg: "rgba(192, 132, 252, 0.12)",
		ThemeBorder: "rgba(192, 132, 252, 0.45)", ThemeGlow: "rgba(192, 132, 252, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/baiak-idle-bot/", Order: 10,
	},
	"luxidle": {
		Key: "luxidle", Icon: "✨", Name: "Lux Idle Multi-Bot", Game: "Lux Idle • Demacia Idle RPG",
		Port: 3012, Theme: "#f59e0b", ThemeBg: "rgba(245, 158, 11, 0.12)",
		ThemeBorder: "rgba(245, 158, 11, 0.45)", ThemeGlow: "rgba(245, 158, 11, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/luxidle-bot/", Order: 12,
	},
	"farmricochet": {
		Key: "farmricochet", Icon: "🌾", Name: "Farm Ricochet Bot", Game: "Farm Ricochet • Web Canvas RPG",
		Port: 3013, Theme: "#9bd26a", ThemeBg: "rgba(155, 210, 106, 0.12)",
		ThemeBorder: "rgba(155, 210, 106, 0.45)", ThemeGlow: "rgba(155, 210, 106, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/farmricochet-bot/", Order: 13,
	},
	"kickminer": {
		Key: "kickminer", Icon: "📺", Name: "Kick Miner Go", Game: "Kick.com • Stream Points & Drops",
		Port: 3009, Theme: "#53fc18", ThemeBg: "rgba(83, 252, 24, 0.12)",
		ThemeBorder: "rgba(83, 252, 24, 0.45)", ThemeGlow: "rgba(83, 252, 24, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/kick-miner-go/", Order: 14,
		IsStream: true,
	},
	"twitchminer": {
		Key: "twitchminer", Icon: "🟣", Name: "Twitch Miner Go", Game: "Twitch.tv • Channel Points & Drops",
		Port: 3010, Theme: "#a855f7", ThemeBg: "rgba(168, 85, 247, 0.12)",
		ThemeBorder: "rgba(168, 85, 247, 0.45)", ThemeGlow: "rgba(168, 85, 247, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/twitch-miner-go/", Order: 15,
		IsStream: true,
	},
	"pocketweb": {
		Key: "pocketweb", Icon: "🌐", Name: "AI-Go-Browser (Navegador)", Game: "AI-Go-Browser • Navegador de IA em Go",
		Port: 8888, Theme: "#3b82f6", ThemeBg: "rgba(59, 130, 246, 0.12)",
		ThemeBorder: "rgba(59, 130, 246, 0.45)", ThemeGlow: "rgba(59, 130, 246, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/ai-go-browser/", Order: 16,
		IsService: true,
	},
	"aigobrowser": {
		Key: "aigobrowser", Icon: "🌐", Name: "AI-Go-Browser (Navegador)", Game: "AI-Go-Browser • Navegador de IA em Go",
		Port: 8888, Theme: "#3b82f6", ThemeBg: "rgba(59, 130, 246, 0.12)",
		ThemeBorder: "rgba(59, 130, 246, 0.45)", ThemeGlow: "rgba(59, 130, 246, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/ai-go-browser/", Order: 16,
		IsService: true,
	},
	"kaggle": {
		Key: "kaggle", Icon: "🧠", Name: "Kaggle Training Watcher", Game: "Kaggle • Autonomous ML Engine",
		Port: 0, Theme: "#06b6d4", ThemeBg: "rgba(6, 182, 212, 0.12)",
		ThemeBorder: "rgba(6, 182, 212, 0.45)", ThemeGlow: "rgba(6, 182, 212, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/agent-go/", Order: 17,
		IsService: true,
	},
	"telegram": {
		Key: "telegram", Icon: "💬", Name: "Telegram Relay Agent", Game: "Telegram • Bot Command Center",
		Port: 0, Theme: "#0284c7", ThemeBg: "rgba(2, 132, 199, 0.12)",
		ThemeBorder: "rgba(2, 132, 199, 0.45)", ThemeGlow: "rgba(2, 132, 199, 0.25)",
		DashboardURL: "https://kbelludoo.github.io/agent-go/", Order: 18,
		IsService: true,
	},
	"lichess": {
		Key: "lichess", Icon: "♟️", Name: "Lichess Bot (Berserk)", Game: "Lichess.org • Go + Stockfish NNUE",
		Port: 3016, Theme: "#ffffff", ThemeBg: "rgba(255, 255, 255, 0.12)",
		ThemeBorder: "rgba(255, 255, 255, 0.45)", ThemeGlow: "rgba(255, 255, 255, 0.25)",
		DashboardURL: "https://lichess.org/@/kbelludoo", Order: 19,
		IsService: true,
	},
}

// GetMeta returns known metadata or automatically generates default presentation for new bots
func GetMeta(key string, b BotData) BotMeta {
	if m, ok := knownMetas[key]; ok {
		if b.Port > 0 {
			m.Port = b.Port
		}
		if b.DashboardURL != "" && key != "pocketweb" && key != "aigobrowser" {
			m.DashboardURL = b.DashboardURL
		}
		return m
	}

	name := b.Name
	if name == "" {
		name = strings.Title(key) + " Bot"
	}
	game := b.Game
	if game == "" {
		game = strings.Title(key) + " • Native Go Engine"
	}

	dash := b.DashboardURL
	if dash == "" {
		dash = fmt.Sprintf("https://kbelludoo.github.io/%s-bot/", key)
	}

	return BotMeta{
		Key:          key,
		Icon:         "🤖",
		Name:         name,
		Game:         game,
		Port:         b.Port,
		Theme:        "#00f0ff",
		ThemeBg:      "rgba(0, 240, 255, 0.12)",
		ThemeBorder:  "rgba(0, 240, 255, 0.45)",
		ThemeGlow:    "rgba(0, 240, 255, 0.25)",
		DashboardURL: dash,
		Order:        99,
	}
}
