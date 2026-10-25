package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"bot-template-go/hubview"
)

func main() {
	statusPath := flag.String("status", "docs/status.json", "Caminho para o status.json")
	outPath := flag.String("out", "docs/index.html", "Caminho de saída para index.html")
	serveAddr := flag.String("serve", "", "Porta para servir via HTTP nativo em Go (ex: :8080)")
	flag.Parse()

	// Função que lê o JSON e gera o HTML usando Go
	buildPage := func() (string, error) {
		raw, err := os.ReadFile(*statusPath)
		if err != nil {
			return "", fmt.Errorf("erro lendo %s: %w", *statusPath, err)
		}

		var data hubview.HubData
		if err := json.Unmarshal(raw, &data); err != nil {
			return "", fmt.Errorf("erro decodificando json: %w", err)
		}

		html, err := hubview.BuildHTML(data)
		if err != nil {
			return "", fmt.Errorf("erro construindo html: %w", err)
		}

		return html, nil
	}

	// 1. Gera o arquivo estático
	htmlContent, err := buildPage()
	if err != nil {
		log.Fatalf("Falha na geração: %v", err)
	}

	if err := os.WriteFile(*outPath, []byte(htmlContent), 0644); err != nil {
		log.Fatalf("Falha ao salvar %s: %v", *outPath, err)
	}
	fmt.Printf("✅ Sucesso! Página gerada em Go com sucesso: %s (%d bytes)\n", *outPath, len(htmlContent))

	// 2. Se a flag -serve for informada, sobe um servidor web Go
	if *serveAddr != "" {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/status.json" {
				http.ServeFile(w, r, *statusPath)
				return
			}
			dynamicHTML, err := buildPage()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(dynamicHTML))
		})

		fmt.Printf("🌐 Servidor Go Web rodando em http://localhost%s\n", *serveAddr)
		if err := http.ListenAndServe(*serveAddr, nil); err != nil {
			log.Fatalf("Erro no servidor: %v", err)
		}
	}
}
