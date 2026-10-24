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
		accs := b.Status.Accounts

		isActive := len(accs) > 0
		if isActive {
			activeCount++
		}

		var totalGold int64
		for _, acc := range accs {
			totalGold += acc.Gold
		}

		summary := fmt.Sprintf("💰 %s gp • RAM: %.1f MB", formatNumber(totalGold), b.Status.RAMSysMB)
		badge := "OFFLINE"
		if isActive {
			badge = fmt.Sprintf("%d/%d ATIVO", len(accs), len(accs))
			if accs[0].Status != "" {
				badge = strings.ToUpper(accs[0].Status)
			}
		}

		viewBots = append(viewBots, ViewBot{
			Meta:        meta,
			Bot:         b,
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
