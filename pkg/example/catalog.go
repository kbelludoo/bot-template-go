package example

import (
	_ "embed"
	"encoding/json"
	"log"
	"sync"

	"bot-template-go/pkg/subsystem"
)

//go:embed data/areas.json
var areasJSON []byte

var (
	catalogOnce sync.Once
	areasList   []subsystem.AreaCandidate
	solver      *subsystem.RouteSolver
)

func InitCatalog() {
	catalogOnce.Do(func() {
		if len(areasJSON) > 0 {
			if err := json.Unmarshal(areasJSON, &areasList); err != nil {
				log.Printf("[example] Erro ao carregar areas.json: %v", err)
			}
		}
		if len(areasList) == 0 {
			// Fallback com áreas básicas
			areasList = []subsystem.AreaCandidate{
				{ID: "area_iniciante", Name: "Área dos Ratos", MinLevel: 1, MaxLevel: 15, BaseExp: 35, KillsPerHour: 450},
				{ID: "area_floresta", Name: "Floresta dos Elfos", MinLevel: 15, MaxLevel: 40, BaseExp: 120, KillsPerHour: 380},
				{ID: "area_caverna", Name: "Caverna dos Dragões", MinLevel: 40, MaxLevel: 100, BaseExp: 550, KillsPerHour: 250},
			}
		}
		solver = subsystem.NewRouteSolver(areasList)
	})
}

func GetRouteSolver() *subsystem.RouteSolver {
	InitCatalog()
	return solver
}
