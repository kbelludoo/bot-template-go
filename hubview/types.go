package hubview

import (
	"encoding/json"
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

// BotStatusData represents live runtime stats of a bot
type BotStatusData struct {
	Uptime     string       `json:"uptime"`
	Goroutines int          `json:"goroutines"`
	RAMAllocMB float64      `json:"ramAllocMB"`
	RAMSysMB   float64      `json:"ramSysMB"`
	Accounts   AccountsList `json:"accounts"`
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
	Extra       map[string]any `json:"extra"`
}
