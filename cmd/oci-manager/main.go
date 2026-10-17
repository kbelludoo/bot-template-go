package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	Tenancy     string
	User        string
	Fingerprint string
	Region      string
	KeyFile     string
	VNICID      string
}

func parsePrivateKey(keyBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}
	if rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return rsaKey, nil
	}
	return nil, fmt.Errorf("unable to parse private key as PKCS#1 or PKCS#8")
}

func signAndDo(cfg *Config, privKey *rsa.PrivateKey, method, path string, bodyBytes []byte) (*http.Response, []byte, error) {
	host := fmt.Sprintf("iaas.%s.oraclecloud.com", cfg.Region)
	url := fmt.Sprintf("https://%s%s", host, path)

	var bodyReader io.Reader
	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, err
	}

	dateHeader := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("Date", dateHeader)
	req.Header.Set("Host", host)

	var headersList string
	var signingParts []string

	target := fmt.Sprintf("(request-target): %s %s", strings.ToLower(method), path)
	datePart := fmt.Sprintf("date: %s", dateHeader)
	hostPart := fmt.Sprintf("host: %s", host)

	if len(bodyBytes) > 0 {
		h := sha256.Sum256(bodyBytes)
		contentSHA := base64.StdEncoding.EncodeToString(h[:])
		contentLength := fmt.Sprintf("%d", len(bodyBytes))
		contentType := "application/json"

		req.Header.Set("x-content-sha256", contentSHA)
		req.Header.Set("Content-Length", contentLength)
		req.Header.Set("Content-Type", contentType)

		headersList = "(request-target) date host x-content-sha256 content-type content-length"
		signingParts = []string{
			target,
			datePart,
			hostPart,
			fmt.Sprintf("x-content-sha256: %s", contentSHA),
			fmt.Sprintf("content-type: %s", contentType),
			fmt.Sprintf("content-length: %s", contentLength),
		}
	} else {
		headersList = "(request-target) date host"
		signingParts = []string{target, datePart, hostPart}
	}

	signingString := strings.Join(signingParts, "\n")
	hashed := sha256.Sum256([]byte(signingString))
	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to sign: %w", err)
	}

	sigB64 := base64.StdEncoding.EncodeToString(sigBytes)
	keyID := fmt.Sprintf("%s/%s/%s", cfg.Tenancy, cfg.User, cfg.Fingerprint)
	authHeader := fmt.Sprintf(`Signature version="1",keyId="%s",algorithm="rsa-sha256",headers="%s",signature="%s"`,
		keyID, headersList, sigB64)

	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	resBody, _ := io.ReadAll(resp.Body)
	return resp, resBody, nil
}

type VNICResponse struct {
	ID            string `json:"id"`
	CompartmentID string `json:"compartmentId"`
	SubnetID      string `json:"subnetId"`
	DisplayName   string `json:"displayName"`
	PublicIP      string `json:"publicIp"`
}

type PrivateIP struct {
	ID              string `json:"id"`
	IPAddress       string `json:"ipAddress"`
	PublicIPAddress string `json:"publicIpAddress"`
	IsPrimary       bool   `json:"isPrimary"`
	DisplayName     string `json:"displayName"`
	CompartmentID   string `json:"compartmentId"`
}

func getVNIC(cfg *Config, privKey *rsa.PrivateKey) (*VNICResponse, error) {
	path := fmt.Sprintf("/20160918/vnics/%s", cfg.VNICID)
	resp, body, err := signAndDo(cfg, privKey, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	var vnic VNICResponse
	if err := json.Unmarshal(body, &vnic); err != nil {
		return nil, err
	}
	return &vnic, nil
}

func main() {
	cmd := flag.String("cmd", "list", "Comando a executar: list, add-ip")
	keyPath := flag.String("key", "/home/ubuntu/.oci/oci_api_key.pem", "Caminho da chave privada RSA")
	flag.Parse()

	cfg := &Config{
		Tenancy:     "ocid1.tenancy.oc1..aaaaaaaamwhan3fl6tskduzntwftob3kz5vihatp76azlxz337zniupm22da",
		User:        "ocid1.user.oc1..aaaaaaaalbcwpembhnptbmbx5rgzzizw6grcu4tsxfrwjygoirwn6smyg3wa",
		Fingerprint: "d1:d0:03:39:9c:fe:98:75:68:9b:61:4b:ee:0f:db:40",
		Region:      "us-ashburn-1",
		KeyFile:     *keyPath,
		VNICID:      "ocid1.vnic.oc1.iad.abuwcljteyk6xvxqk7fzlaqo2ocusfbw3oqpaved6gdf5fcrm3lrx6mn4jyq",
	}

	keyData, err := os.ReadFile(cfg.KeyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler chave %s: %v\n", cfg.KeyFile, err)
		os.Exit(1)
	}

	privKey, err := parsePrivateKey(keyData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao parsear chave: %v\n", err)
		os.Exit(1)
	}

	vnic, err := getVNIC(cfg, privKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao obter VNIC: %v\n", err)
		os.Exit(1)
	}

	switch *cmd {
	case "list":
		fmt.Printf("🔍 Consultando VNIC: %s (Compartimento: %s)...\n", vnic.DisplayName, vnic.CompartmentID)
		path := fmt.Sprintf("/20160918/privateIps?vnicId=%s", cfg.VNICID)
		resp, body, err := signAndDo(cfg, privKey, "GET", path, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro na requisição: %v\n", err)
			os.Exit(1)
		}
		if resp.StatusCode >= 300 {
			fmt.Printf("HTTP %d: %s\n", resp.StatusCode, string(body))
			os.Exit(1)
		}

		var privateIPs []PrivateIP
		_ = json.Unmarshal(body, &privateIPs)

		// Consultar IPs públicos associados
		pubPath := fmt.Sprintf("/20160918/publicIps?compartmentId=%s&scope=REGION", vnic.CompartmentID)
		_, pubBody, _ := signAndDo(cfg, privKey, "GET", pubPath, nil)
		var publicIPs []struct {
			ID          string `json:"id"`
			IPAddress   string `json:"ipAddress"`
			PrivateIPID string `json:"privateIpId"`
			Lifetime    string `json:"lifetime"`
		}
		_ = json.Unmarshal(pubBody, &publicIPs)
		pubMap := make(map[string]string)
		for _, p := range publicIPs {
			pubMap[p.PrivateIPID] = p.IPAddress
		}

		fmt.Printf("✅ Autenticado com sucesso! Total de IPs Privados na VNIC: %d\n", len(privateIPs))
		for i, pip := range privateIPs {
			pub := pip.PublicIPAddress
			if pub == "" {
				pub = pubMap[pip.ID]
			}
			if pub == "" {
				pub = "Nenhum (Somente Privado)"
			}
			prim := ""
			if pip.IsPrimary {
				prim = " [PRIMÁRIO]"
			}
			fmt.Printf("  [%d] Privado: %-14s | Público: %-16s | ID: %s%s\n", i+1, pip.IPAddress, pub, pip.ID, prim)
		}

	case "add-ip":
		fmt.Printf("🚀 [1/3] Criando IP Privado Secundário na VNIC %s...\n", vnic.DisplayName)
		path := "/20160918/privateIps"
		payload := map[string]any{
			"vnicId": cfg.VNICID,
		}
		bodyBytes, _ := json.Marshal(payload)

		resp, body, err := signAndDo(cfg, privKey, "POST", path, bodyBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar private IP: %v\n", err)
			os.Exit(1)
		}
		if resp.StatusCode >= 300 {
			fmt.Printf("❌ Falha na criação do IP Privado (HTTP %d):\n%s\n", resp.StatusCode, string(body))
			os.Exit(1)
		}

		var created PrivateIP
		_ = json.Unmarshal(body, &created)
		fmt.Printf("✅ IP Privado Criado: %s (OCID: %s)\n", created.IPAddress, created.ID)

		// 2. Criar e associar IP Público Reservado
		fmt.Println("🚀 [2/3] Solicitando e associando IP Público Reservado na Oracle...")
		pubPath := "/20160918/publicIps"
		pubPayload := map[string]any{
			"compartmentId": vnic.CompartmentID,
			"lifetime":      "RESERVED",
			"privateIpId":   created.ID,
			"displayName":   fmt.Sprintf("reserved-public-%s", created.IPAddress),
		}
		pubBytes, _ := json.Marshal(pubPayload)
		pubResp, pubResBody, pubErr := signAndDo(cfg, privKey, "POST", pubPath, pubBytes)
		if pubErr != nil || pubResp.StatusCode >= 300 {
			fmt.Printf("⚠️ Aviso na criação de IP Público (HTTP %d): %s\n", pubResp.StatusCode, string(pubResBody))
		} else {
			var pubCreated struct {
				IPAddress string `json:"ipAddress"`
			}
			_ = json.Unmarshal(pubResBody, &pubCreated)
			fmt.Printf("🎉 IP PÚBLICO CRIADO COM SUCESSO: %s\n", pubCreated.IPAddress)
		}

		// 3. Adicionar interface na VPS Linux local
		fmt.Printf("🚀 [3/3] Configurando interface Linux ens3 com IP %s/24...\n", created.IPAddress)
		ipCmd := exec.Command("sudo", "ip", "addr", "add", fmt.Sprintf("%s/24", created.IPAddress), "dev", "ens3")
		if err := ipCmd.Run(); err != nil {
			fmt.Printf("⚠️ Interface local: %v (pode já estar adicionada ou exigir netplan)\n", err)
		} else {
			fmt.Printf("✅ Rota e interface ens3 configuradas para navegar imediatamente!\n")
		}

	case "attach-public":
		pipID := flag.Arg(0)
		if pipID == "" {
			fmt.Println("Uso: oci-manager -cmd=attach-public <OCID_DO_PRIVATE_IP>")
			os.Exit(1)
		}
		fmt.Printf("🚀 Solicitando IP Público Reservado para Private IP %s...\n", pipID)
		pubPath := "/20160918/publicIps"
		pubPayload := map[string]any{
			"compartmentId": vnic.CompartmentID,
			"lifetime":      "RESERVED",
			"privateIpId":   pipID,
			"displayName":   fmt.Sprintf("reserved-public-auto"),
		}
		pubBytes, _ := json.Marshal(pubPayload)
		pubResp, pubResBody, pubErr := signAndDo(cfg, privKey, "POST", pubPath, pubBytes)
		if pubErr != nil || pubResp.StatusCode >= 300 {
			fmt.Printf("❌ Erro na criação de IP Público (HTTP %d): %s\n", pubResp.StatusCode, string(pubResBody))
			var pubCreated struct {
				IPAddress string `json:"ipAddress"`
			}
			_ = json.Unmarshal(pubResBody, &pubCreated)
			fmt.Printf("🎉 IP PÚBLICO CRIADO COM SUCESSO: %s\n", pubCreated.IPAddress)
		}

	case "test-limit":
		maxAttempts := 15
		fmt.Printf("🎯 Testando limite de IPs públicos da Oracle (tentando criar até %d novos IPs)...\n", maxAttempts)
		successCount := 0
		for i := 1; i <= maxAttempts; i++ {
			fmt.Printf("\n--- [Lote #%d] Solicitando novo IP... ---\n", i)
			// 1. Criar private IP
			path := "/20160918/privateIps"
			payload := map[string]any{"vnicId": cfg.VNICID}
			bodyBytes, _ := json.Marshal(payload)
			resp, body, err := signAndDo(cfg, privKey, "POST", path, bodyBytes)
			if err != nil || resp.StatusCode >= 300 {
				fmt.Printf("🛑 Limite de IPs Privados na VNIC atingido (HTTP %d): %s\n", resp.StatusCode, string(body))
				break
			}
			var created PrivateIP
			_ = json.Unmarshal(body, &created)

			// 2. Criar public IP
			pubPath := "/20160918/publicIps"
			pubPayload := map[string]any{
				"compartmentId": vnic.CompartmentID,
				"lifetime":      "RESERVED",
				"privateIpId":   created.ID,
				"displayName":   fmt.Sprintf("reserved-public-%s", created.IPAddress),
			}
			pubBytes, _ := json.Marshal(pubPayload)
			pubResp, pubResBody, pubErr := signAndDo(cfg, privKey, "POST", pubPath, pubBytes)
			if pubErr != nil || pubResp.StatusCode >= 300 {
				fmt.Printf("🛑 LIMITE DE IPS PÚBLICOS ATINGIDO PELA ORACLE!\nMotivo (HTTP %d): %s\n", pubResp.StatusCode, string(pubResBody))
				// Limpar private IP criado sem public IP
				delPath := fmt.Sprintf("/20160918/privateIps/%s", created.ID)
				_, _, _ = signAndDo(cfg, privKey, "DELETE", delPath, nil)
				break
			}

			var pubCreated struct {
				IPAddress string `json:"ipAddress"`
			}
			_ = json.Unmarshal(pubResBody, &pubCreated)
			fmt.Printf("🎉 IP #%d CRIADO COM SUCESSO!\n   Privado: %s ➜ Público: %s\n", successCount+1, created.IPAddress, pubCreated.IPAddress)

			// 3. Vincular interface local
			_ = exec.Command("sudo", "ip", "addr", "add", fmt.Sprintf("%s/24", created.IPAddress), "dev", "ens3").Run()
			successCount++
			time.Sleep(1 * time.Second)
		}
		fmt.Printf("\n🏁 Teste de limite finalizado! Novos IPs criados nesta rodada: %d\n", successCount)

	default:
		fmt.Printf("Comando desconhecido: %s. Use -cmd=list, -cmd=add-ip, -cmd=test-limit ou -cmd=attach-public\n", *cmd)
	}
}
