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

func TestKickMinerUnifiedAccounts(t *testing.T) {
	raw := `{
		"accounts": {
			"miner_1": {
				"activeStreams": 3,
				"streamers": {
					"gaules": {
						"streamer": "gaules",
						"points": 9330,
						"isLive": true,
						"watching": true
					},
					"adinross": {
						"streamer": "adinross",
						"points": 1600,
						"isLive": false,
						"watching": false
					}
				}
			}
		}
	}`

	var status BotStatusData
	if err := json.Unmarshal([]byte(raw), &status); err != nil {
		t.Fatalf("falha ao deserializar json kick: %v", err)
	}

	accs := status.GetUnifiedAccounts()
	if len(accs) != 1 {
		t.Fatalf("esperava 1 conta kick, obteve %d", len(accs))
	}

	if accs[0].Gold != 10930 {
		t.Errorf("esperava 10930 pontos no Kick, obteve %d", accs[0].Gold)
	}
	if !accs[0].IsStream {
		t.Errorf("esperava IsStream true para Kick")
	}
	if accs[0].Kills != 1 {
		t.Errorf("esperava 1 canal ao vivo, obteve %d", accs[0].Kills)
	}
}

func TestBuildHTMLStreamDistinction(t *testing.T) {
	data := HubData{
		Timestamp: "2026-10-24T12:00:00Z",
		Bots: map[string]BotData{
			"kickminer": {
				Port: 3009,
				Name: "Kick Miner Go",
				Status: BotStatusData{
					Accounts: json.RawMessage(`{"miner_1":{"activeStreams":1,"streamers":{"gaules":{"points":5000,"isLive":true}}}}`),
				},
			},
		},
	}

	html, err := BuildHTML(data)
	if err != nil {
		t.Fatalf("BuildHTML falhou: %v", err)
	}

	if !strings.Contains(html, "pontos") {
		t.Errorf("html gerado para stream miner deve conter 'pontos'")
	}
	if strings.Contains(html, "gp") && strings.Contains(html, "card-kickminer") {
		// Verify kickminer card does not show gp
		idx := strings.Index(html, "card-kickminer")
		slice := html[idx : idx+500]
		if strings.Contains(slice, "gp") {
			t.Errorf("card do kickminer não deve conter 'gp': %s", slice)
		}
	}
}
