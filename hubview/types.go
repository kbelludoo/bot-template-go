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

// FlexibleString unmarshals both numbers and strings seamlessly
type FlexibleString string

func (fs *FlexibleString) UnmarshalJSON(b []byte) error {
	var n float64
	if err := json.Unmarshal(b, &n); err == nil {
		*fs = FlexibleString(fmt.Sprintf("%d%%", int(n)))
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*fs = FlexibleString(s)
		return nil
	}
	*fs = ""
	return nil
}

func (fs FlexibleString) String() string {
	return string(fs)
}

// KickStreamerEntry represents individual streamer channel monitored in Kick Miner
type KickStreamerEntry struct {
	Streamer  string `json:"streamer"`
	IsLive    bool   `json:"isLive"`
	Watching  bool   `json:"watching"`
	Points    int64  `json:"points"`
	WatchTime string `json:"watchTime"`
}

// KickAccountEntry represents an account entry in Kick Miner
type KickAccountEntry struct {
	ActiveStreams int                          `json:"activeStreams"`
	Streamers     map[string]KickStreamerEntry `json:"streamers"`
}

// BaiakBotEntry represents individual bot in Baiak Idle fleet
type BaiakBotEntry struct {
	Label     string         `json:"label"`
	Name      string         `json:"name"`
	Online    bool           `json:"online"`
	Reachable bool           `json:"reachable"`
	Status    BaiakBotStatus `json:"status"`
}

// BaiakRatesInfo holds live exp and gold rates
type BaiakRatesInfo struct {
	GoldPerHour float64 `json:"goldPerHour"`
	XpPerHour   float64 `json:"xpPerHour"`
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
	UptimeSec   int                `json:"uptimeSec"`
	Hunt        BaiakHuntInfo      `json:"hunt"`
	Telemetry   BaiakTelemetryInfo `json:"telemetry"`
	Rates       BaiakRatesInfo     `json:"rates"`
}

// BaiakHuntInfo represents hunt target in Baiak
type BaiakHuntInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BaiakTelemetryInfo represents telemetry metrics in Baiak
type BaiakTelemetryInfo struct {
	Gold    int64          `json:"gold"`
	Kills   int            `json:"kills"`
	Level   int            `json:"level"`
	Stamina FlexibleString `json:"stamina"`
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
	Uptime        string                       `json:"uptime"`
	Goroutines    int                          `json:"goroutines"`
	RAMAllocMB    float64                      `json:"ramAllocMB"`
	RAMSysMB      float64                      `json:"ramSysMB"`
	RAMMB         float64                      `json:"ramMb"`
	TotalStreams  int                          `json:"totalStreams"`
	Accounts      json.RawMessage              `json:"accounts"`
	Bots          map[string]BaiakBotEntry     `json:"bots"`
	Statuses      map[string]TwitchStatusEntry `json:"statuses"`
}

// AccountData represents a single character runner or stream harvester
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
	GoldPerHour float64        `json:"goldPerHour,omitempty"`
	Stamina     FlexibleString `json:"stamina,omitempty"`
	IsStream    bool           `json:"isStream,omitempty"`
	Extra       map[string]any `json:"extra,omitempty"`
}

// GetUnifiedAccounts returns a normalized slice of AccountData across all game and stream bot types
func (s BotStatusData) GetUnifiedAccounts() []AccountData {
	// 1. Kick Miner check (accounts as map[string]KickAccountEntry)
	if len(s.Accounts) > 0 {
		var kickMap map[string]KickAccountEntry
		if err := json.Unmarshal(s.Accounts, &kickMap); err == nil && len(kickMap) > 0 {
			hasKick := false
			for _, v := range kickMap {
				if len(v.Streamers) > 0 || v.ActiveStreams > 0 {
					hasKick = true
					break
				}
			}
			if hasKick {
				var keys []string
				for k := range kickMap {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				var res []AccountData
				for _, k := range keys {
					entry := kickMap[k]
					var totalPts int64
					var liveCount int
					var liveNames []string

					var sNames []string
					for sName := range entry.Streamers {
						sNames = append(sNames, sName)
					}
					sort.Strings(sNames)

					for _, sName := range sNames {
						st := entry.Streamers[sName]
						totalPts += st.Points
						if st.IsLive {
							liveCount++
							if len(liveNames) < 3 {
								liveNames = append(liveNames, st.Streamer)
							}
						}
					}

					area := fmt.Sprintf("%d canais ativos", entry.ActiveStreams)
					if entry.ActiveStreams == 1 {
						area = "1 canal ativo"
					}
					if len(liveNames) > 0 {
						area = fmt.Sprintf("Ao vivo: %s (+%d)", strings.Join(liveNames, ", "), liveCount)
					}

					name := "Conta Kick #1"
					if k != "miner_1" && k != "" {
						name = k
					}

					res = append(res, AccountData{
						ID:          k,
						Name:        name,
						Level:       0,
						Gold:        totalPts,
						Kills:       liveCount,
						Vocation:    "Kick Stream",
						CurrentArea: area,
						Status:      "Minerando",
						IsStream:    true,
					})
				}
				return res
			}
		}

		// Standard slice []AccountData (with FlexibleString for stamina like in Lux Idle)
		var list []AccountData
		if err := json.Unmarshal(s.Accounts, &list); err == nil && len(list) > 0 {
			return list
		}

		// Standard map map[string]AccountData
		var m map[string]AccountData
		if err := json.Unmarshal(s.Accounts, &m); err == nil && len(m) > 0 {
			var res []AccountData
			for k, v := range m {
				if v.Name == "" {
					v.Name = k
				}
				res = append(res, v)
			}
			return res
		}
	}

	// 2. Baiak Idle fleet check
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
				DurationSec: b.Status.UptimeSec,
				ExpPerHour:  b.Status.Rates.XpPerHour,
				GoldPerHour: b.Status.Rates.GoldPerHour,
			})
		}
		return res
	}

	// 3. Twitch Miner check
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
			var liveCount int
			for _, stream := range st.Streamers {
				if stream.Watching {
					watching = append(watching, stream.DisplayName)
				}
				if stream.IsLive {
					liveCount++
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
				Level:       0,
				Gold:        totalPts,
				Kills:       liveCount,
				Vocation:    "Twitch Drops",
				CurrentArea: area,
				Status:      "Online",
				IsStream:    true,
			})
		}
		return res
	}

	// 4. Standby process fallback (daemon awaiting match/session)
	if s.Uptime != "" || s.RAMSysMB > 0 || s.Goroutines > 0 {
		return []AccountData{
			{
				ID:          "standby-1",
				Name:        "Sessão Conectada",
				Level:       1,
				Vocation:    "Operação Contínua",
				CurrentArea: "Monitorando em Standby",
				Status:      "Standby",
				DurationSec: 0,
			},
		}
	}

	return nil
}
