package hubview

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBaiakUnifiedAccounts(t *testing.T) {
	raw := `{
		"bots": {
			"acc-1": {
				"label": "Conta 1",
				"online": true,
				"status": {
					"level": 373,
					"vocation": "paladin",
					"currentArea": "Vexclaw",
					"gold": 106042234,
					"kills": 6244,
					"status": "hunting",
					"telemetry": {
						"stamina": "36h 39m (87%)"
					}
				}
			},
			"acc-2": {
				"label": "Conta 2",
				"online": true,
				"status": {
					"level": 260,
					"vocation": "knight",
					"currentArea": "Stone Refiner",
					"gold": 191438184,
					"kills": 39219,
					"treino": true,
					"status": "training",
					"telemetry": {
						"stamina": "22h 30m (54%)"
					}
				}
			}
		}
	}`

	var status BotStatusData
	if err := json.Unmarshal([]byte(raw), &status); err != nil {
		t.Fatalf("falha ao deserializar json: %v", err)
	}

	accs := status.GetUnifiedAccounts()
	if len(accs) != 2 {
		t.Fatalf("esperava 2 contas, obteve %d", len(accs))
	}

	if accs[0].Name != "Conta 1" || accs[0].Level != 373 || accs[0].Vocation != "paladin" {
		t.Errorf("conta 1 incorreta: %+v", accs[0])
	}
	if accs[0].Stamina != "36h 39m (87%)" {
		t.Errorf("stamina da conta 1 incorreta: %s", accs[0].Stamina)
	}

	if accs[1].Name != "Conta 2" || accs[1].Status != "Treino" {
		t.Errorf("conta 2 incorreta: %+v", accs[1])
	}
}

func TestBuildHTMLContainsBaiak(t *testing.T) {
	data := HubData{
		Timestamp: "2026-10-24T12:00:00Z",
		VPS: VPSInfo{
			RAMTotalMB: 1024,
			RAMUsedMB:  512,
		},
		Bots: map[string]BotData{
			"baiak": {
				Port:         8090,
				Name:         "Baiak Idle Multi-Bot",
				Game:         "Baiak Idle • Open Tibia Idle",
				DashboardURL: "https://kbelludoo.github.io/baiak-idle-bot/",
				Status: BotStatusData{
					Bots: map[string]BaiakBotEntry{
						"acc-1": {
							Label:  "Conta 1",
							Online: true,
							Status: BaiakBotStatus{
								Level:       373,
								Vocation:    "paladin",
								CurrentArea: "Vexclaw",
								Gold:        100000000,
								Kills:       5000,
								Status:      "hunting",
								Telemetry: BaiakTelemetryInfo{
									Stamina: "36h 39m (87%)",
								},
							},
						},
					},
				},
			},
		},
	}

	html, err := BuildHTML(data)
	if err != nil {
		t.Fatalf("BuildHTML falhou: %v", err)
	}

	if !strings.Contains(html, "Conta 1 (Nv. 373)") {
		t.Errorf("html gerado não contém a conta 1 de baiak")
	}
	if !strings.Contains(html, "36h 39m (87%)") {
		t.Errorf("html gerado não contém a stamina da conta 1")
	}
	if !strings.Contains(html, "https://kbelludoo.github.io/baiak-idle-bot/") {
		t.Errorf("html gerado não contém o dashboard correto de baiak")
	}
}
