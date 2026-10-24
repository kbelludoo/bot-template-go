package hubview

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// HubData represents the complete cluster telemetry payload
type HubData struct {
	Timestamp string             `json:"timestamp"`
	Summary   HubSummary         `json:"summary"`
	VPS       VPSInfo            `json:"vps"`
	Bots      map[string]BotData `json:"bots"`
}

// HubSummary provides high-level bot counters
type HubSummary struct {
	ActiveBots int `json:"activeBots"`
	TotalBots  int `json:"totalBots"`
}

// VPSInfo contains hardware metrics from the host
type VPSInfo struct {
	RAMTotalMB     int    `json:"ramTotalMb"`
	RAMUsedMB      int    `json:"ramUsedMb"`
	RAMAvailableMB int    `json:"ramAvailableMb"`
	RAMPercent     int    `json:"ramPercent"`
	SwapTotalMB    int    `json:"swapTotalMb"`
	SwapUsedMB     int    `json:"swapUsedMb"`
	Load           string `json:"load"`
	Uptime         string `json:"uptime"`
}

// BotData contains individual game bot metadata and telemetry
type BotData struct {
	Port         int           `json:"port"`
	Name         string        `json:"name"`
	Game         string        `json:"game"`
	Engine       string        `json:"engine"`
	DashboardURL string        `json:"dashboardUrl"`
	Status       BotStatusData `json:"status"`
}

// AccountsList handles both JSON arrays []AccountData and JSON maps map[string]AccountData
type AccountsList []AccountData

func (a *AccountsList) UnmarshalJSON(data []byte) error {
	var list []AccountData
	if err := json.Unmarshal(data, &list); err == nil {
		*a = list
		return nil
	}
	var m map[string]AccountData
	if err := json.Unmarshal(data, &m); err == nil {
		var res []AccountData
		for k, v := range m {
			if v.Name == "" {
				v.Name = k
			}
			res = append(res, v)
		}
		*a = res
		return nil
	}
	*a = []AccountData{}
	return nil
}

// BaiakBotEntry represents individual bot in Baiak Idle fleet
type BaiakBotEntry struct {
	Label     string         `json:"label"`
	Name      string         `json:"name"`
	Online    bool           `json:"online"`
	Reachable bool           `json:"reachable"`
	Status    BaiakBotStatus `json:"status"`
}

// BaiakBotStatus represents inner state of a Baiak character
type BaiakBotStatus struct {
	CurrentArea string             `json:"currentArea"`
	Gold        int64              `json:"gold"`
	Kills       int                `json:"kills"`
	Level       int                `json:"level"`
	Online      bool               `json:"online"`
	Status      string             `json:"status"`
	Treino      bool               `json:"treino"`
	Vocation    string             `json:"vocation"`
	Hunt        BaiakHuntInfo      `json:"hunt"`
	Telemetry   BaiakTelemetryInfo `json:"telemetry"`
}

// BaiakHuntInfo represents hunt target in Baiak
type BaiakHuntInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BaiakTelemetryInfo represents telemetry metrics in Baiak
type BaiakTelemetryInfo struct {
	Gold    int64  `json:"gold"`
	Kills   int    `json:"kills"`
	Level   int    `json:"level"`
	Stamina string `json:"stamina"`
}

// TwitchStatusEntry represents a miner account in Twitch Miner
type TwitchStatusEntry struct {
	UserID    string                         `json:"userId"`
	Username  string                         `json:"username"`
	Streamers map[string]TwitchStreamerEntry `json:"streamers"`
}

// TwitchStreamerEntry represents a streamer channel monitored by Twitch Miner
type TwitchStreamerEntry struct {
	DisplayName string `json:"displayName"`
	IsLive      bool   `json:"isLive"`
	Watching    bool   `json:"watching"`
	Points      int64  `json:"points"`
	Title       string `json:"title"`
	Viewers     int    `json:"viewers"`
}

// BotStatusData represents live runtime stats of a bot
type BotStatusData struct {
	Uptime     string                       `json:"uptime"`
	Goroutines int                          `json:"goroutines"`
	RAMAllocMB float64                      `json:"ramAllocMB"`
	RAMSysMB   float64                      `json:"ramSysMB"`
	Accounts   AccountsList                 `json:"accounts"`
	Bots       map[string]BaiakBotEntry     `json:"bots"`
	Statuses   map[string]TwitchStatusEntry `json:"statuses"`
}

// AccountData represents a single game character or runner
type AccountData struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Level       int            `json:"level"`
	HP          int            `json:"hp"`
	MaxHP       int            `json:"maxHp"`
	Mana        int            `json:"mana"`
	MaxMana     int            `json:"maxMana"`
	Gold        int64          `json:"gold"`
	Kills       int            `json:"kills"`
	Vocation    string         `json:"vocation"`
	CurrentArea string         `json:"currentArea"`
	Status      string         `json:"status"`
	DurationSec int            `json:"durationSec"`
	ExpPerHour  float64        `json:"expPerHour"`
	Stamina     string         `json:"stamina,omitempty"`
	Extra       map[string]any `json:"extra,omitempty"`
}

// GetUnifiedAccounts returns a normalized slice of AccountData across all bot types
func (s BotStatusData) GetUnifiedAccounts() []AccountData {
	if len(s.Accounts) > 0 {
		return s.Accounts
	}

	if len(s.Bots) > 0 {
		var keys []string
		for k := range s.Bots {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var res []AccountData
		for _, k := range keys {
			b := s.Bots[k]
			name := b.Label
			if name == "" {
				name = b.Name
			}
			if name == "" {
				name = k
			}

			statusStr := b.Status.Status
			if b.Status.Treino {
				statusStr = "Treino"
			}
			area := b.Status.CurrentArea
			if area == "" && b.Status.Hunt.Name != "" {
				area = b.Status.Hunt.Name
			}
			stamina := b.Status.Telemetry.Stamina

			gold := b.Status.Gold
			if gold == 0 && b.Status.Telemetry.Gold > 0 {
				gold = b.Status.Telemetry.Gold
			}
			kills := b.Status.Kills
			if kills == 0 && b.Status.Telemetry.Kills > 0 {
				kills = b.Status.Telemetry.Kills
			}
			lvl := b.Status.Level
			if lvl == 0 && b.Status.Telemetry.Level > 0 {
				lvl = b.Status.Telemetry.Level
			}

			res = append(res, AccountData{
				ID:          k,
				Name:        name,
				Level:       lvl,
				Gold:        gold,
				Kills:       kills,
				Vocation:    b.Status.Vocation,
				CurrentArea: area,
				Status:      statusStr,
				Stamina:     stamina,
			})
		}
		return res
	}

	if len(s.Statuses) > 0 {
		var keys []string
		for k := range s.Statuses {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var res []AccountData
		for _, k := range keys {
			st := s.Statuses[k]
			name := st.Username
			if name == "" {
				name = k
			}

			var watching []string
			var totalPts int64
			for _, stream := range st.Streamers {
				if stream.Watching {
					watching = append(watching, stream.DisplayName)
				}
				totalPts += stream.Points
			}

			area := "Minerando canais Twitch"
			if len(watching) > 0 {
				area = fmt.Sprintf("Assistindo %s", strings.Join(watching, ", "))
			}

			res = append(res, AccountData{
				ID:          st.UserID,
				Name:        name,
				Level:       1,
				Gold:        totalPts,
				Vocation:    "Twitch Drops",
				CurrentArea: area,
				Status:      "Online",
			})
		}
		return res
	}

	return nil
}
