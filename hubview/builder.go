package hubview

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"
)

// ViewBot holds computed view-model data for template rendering
type ViewBot struct {
	Meta        BotMeta
	Bot         BotData
	Accounts    []AccountData
	SummaryText string
	BadgeText   string
	IsActive    bool
}

// ViewHub holds the root model passed to the template
type ViewHub struct {
	Timestamp      string
	TimeAgo        string
	ActiveBots     int
	TotalBots      int
	VPS            VPSInfo
	Bots           []ViewBot
	TickerBotsList string
}

// formatNumber converts numbers to human-readable strings (e.g. 1.234 or 45.2k)
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1_000_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%.2fkk", float64(n)/1_000_000.0)
}

// BuildViewHub creates the pre-computed view model from raw telemetry
func BuildViewHub(data HubData) ViewHub {
	var sortedKeys []string
	for k := range data.Bots {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Slice(sortedKeys, func(i, j int) bool {
		mI := GetMeta(sortedKeys[i], data.Bots[sortedKeys[i]])
		mJ := GetMeta(sortedKeys[j], data.Bots[sortedKeys[j]])
		if mI.Order != mJ.Order {
			return mI.Order < mJ.Order
		}
		return sortedKeys[i] < sortedKeys[j]
	})

	var viewBots []ViewBot
	activeCount := 0
	var tickerNames []string

	for _, k := range sortedKeys {
		b := data.Bots[k]
		meta := GetMeta(k, b)
		accs := b.Status.GetUnifiedAccounts()

		isActive := len(accs) > 0
		if isActive {
			activeCount++
		}

		var totalVal int64
		for _, acc := range accs {
			totalVal += acc.Gold
		}

		ramMB := b.Status.RAMSysMB
		if ramMB == 0 && b.Status.RAMMB > 0 {
			ramMB = b.Status.RAMMB
		}

		var summary string
		if meta.IsStream {
			summary = fmt.Sprintf("⭐ %s pontos", formatNumber(totalVal))
			if ramMB > 0 {
				summary = fmt.Sprintf("⭐ %s pontos • RAM: %.1f MB", formatNumber(totalVal), ramMB)
			}
		} else if meta.IsService {
			summary = "⚙️ Serviço Ativo 24/7"
			if ramMB > 0 {
				summary = fmt.Sprintf("⚙️ RAM: %.1f MB • 24/7", ramMB)
			}
		} else {
			summary = fmt.Sprintf("💰 %s gp", formatNumber(totalVal))
			if ramMB > 0 {
				summary = fmt.Sprintf("💰 %s gp • RAM: %.1f MB", formatNumber(totalVal), ramMB)
			}
		}

		badge := "OFFLINE"
		if isActive {
			if meta.IsStream {
				if len(accs) > 0 && accs[0].Kills > 0 {
					badge = fmt.Sprintf("%d AO VIVO", accs[0].Kills)
				} else {
					badge = "MINERANDO"
				}
			} else if meta.IsService {
				if len(accs) > 0 && accs[0].Status != "" {
					badge = strings.ToUpper(accs[0].Status)
				} else {
					badge = "ONLINE"
				}
			} else if len(accs) > 1 {
				badge = fmt.Sprintf("%d/%d ATIVAS", len(accs), len(accs))
			} else if len(accs) == 1 {
				if accs[0].Status != "" {
					badge = strings.ToUpper(accs[0].Status)
				} else {
					badge = "ONLINE"
				}
			}
		}

		viewBots = append(viewBots, ViewBot{
			Meta:        meta,
			Bot:         b,
			Accounts:    accs,
			SummaryText: summary,
			BadgeText:   badge,
			IsActive:    isActive,
		})

		tickerNames = append(tickerNames, strings.ToUpper(meta.Name))
	}

	timeAgo := "Agora há pouco"
	if t, err := time.Parse(time.RFC3339, data.Timestamp); err == nil {
		diff := time.Since(t).Round(time.Second)
		timeAgo = fmt.Sprintf("há %s", diff)
	}

	totalBots := len(data.Bots)
	if data.Summary.TotalBots > totalBots {
		totalBots = data.Summary.TotalBots
	}

	return ViewHub{
		Timestamp:      data.Timestamp,
		TimeAgo:        timeAgo,
		ActiveBots:     activeCount,
		TotalBots:      totalBots,
		VPS:            data.VPS,
		Bots:           viewBots,
		TickerBotsList: strings.Join(tickerNames, " • "),
	}
}

// BuildHTML executes the Go template and returns the complete standalone HTML string
func BuildHTML(data HubData) (string, error) {
	view := BuildViewHub(data)

	tmpl, err := template.New("hub").Funcs(template.FuncMap{
		"formatNumber": formatNumber,
		"css": func(s string) template.CSS {
			return template.CSS(s)
		},
		"formatRate": func(n float64) string {
			if n <= 0 {
				return "Calculando..."
			}
			return "+" + formatNumber(int64(n)) + "/h"
		},
		"formatUptime": func(sec int, fallback string) string {
			if sec > 0 {
				h := sec / 3600
				m := (sec % 3600) / 60
				return fmt.Sprintf("%dh %02dm", h, m)
			}
			if fallback != "" {
				return fallback
			}
			return "24/7 Ativo"
		},
	}).Parse(hubTemplate)
	if err != nil {
		return "", fmt.Errorf("falha ao analisar template do hub: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, view); err != nil {
		return "", fmt.Errorf("falha ao renderizar html do hub: %w", err)
	}

	return buf.String(), nil
}
