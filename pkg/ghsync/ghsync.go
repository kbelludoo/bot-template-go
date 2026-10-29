package ghsync

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const Owner = "kbelludoo"

var (
	tokenCache     string
	tokenCacheOnce sync.Once
	httpClient     = &http.Client{Timeout: 15 * time.Second}
)

// ResolveToken obtém o token do GitHub via env ou via arquivo de configuração do gh CLI
func ResolveToken() string {
	tokenCacheOnce.Do(func() {
		if t := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); t != "" {
			tokenCache = t
			return
		}
		if t := strings.TrimSpace(os.Getenv("GH_TOKEN")); t != "" {
			tokenCache = t
			return
		}

		home, err := os.UserHomeDir()
		if err == nil {
			candidates := []string{
				filepath.Join(home, ".config", "gh", "hosts.yml"),
				filepath.Join(os.Getenv("APPDATA"), "GitHub CLI", "hosts.yml"),
				"/home/ubuntu/.secrets.json",
				filepath.Join(home, ".secrets.json"),
			}
			for _, p := range candidates {
				data, err := os.ReadFile(p)
				if err != nil {
					continue
				}
				content := string(data)
				// Se for YAML do gh
				if strings.Contains(content, "oauth_token:") {
					for _, line := range strings.Split(content, "\n") {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, "oauth_token:") {
							parts := strings.SplitN(line, ":", 2)
							if len(parts) == 2 {
								t := strings.TrimSpace(parts[1])
								t = strings.Trim(t, `"'`)
								if t != "" {
									tokenCache = t
									return
								}
							}
						}
					}
				}
				// Se for JSON de segredos
				if strings.Contains(content, "githubToken") || strings.Contains(content, "ghToken") {
					var raw map[string]string
					if err := json.Unmarshal(data, &raw); err == nil {
						if t, ok := raw["githubToken"]; ok && t != "" {
							tokenCache = t
							return
						}
						if t, ok := raw["ghToken"]; ok && t != "" {
							tokenCache = t
							return
						}
					}
				}
			}
		}
	})
	return tokenCache
}

// GH executa o comando CLI com timeout (fallback legado)
func GH(timeout time.Duration, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("gh %s: %s", strings.Join(args, " "), msg)
	}
	return out.Bytes(), nil
}

func contentURL(repo, path string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", Owner, repo, strings.TrimPrefix(path, "/"))
}

// FileSHA busca o SHA atual do arquivo sem criar subprocessos
func FileSHA(repo, path string) (string, error) {
	token := ResolveToken()
	if token == "" {
		// Fallback para gh CLI
		out, err := GH(20*time.Second, "api", fmt.Sprintf("repos/%s/%s/contents/%s", Owner, repo, path), "--jq", ".sha")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}

	req, err := http.NewRequest("GET", contentURL(repo, path), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "bot-telemetry-go")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http sha error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil // Arquivo ainda não existe
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github api get status %d", resp.StatusCode)
	}

	var res struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.SHA, nil
}

// Content busca o conteúdo decodificado de um arquivo
func Content(repo, path string) ([]byte, error) {
	token := ResolveToken()
	if token == "" {
		out, err := GH(30*time.Second, "api", fmt.Sprintf("repos/%s/%s/contents/%s", Owner, repo, path), "--jq", ".content")
		if err != nil {
			return nil, err
		}
		return decodeBase64(string(out))
	}

	req, err := http.NewRequest("GET", contentURL(repo, path), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "bot-telemetry-go")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http content error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api status %d", resp.StatusCode)
	}

	var res struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return decodeBase64(res.Content)
}

// ContentAt busca o conteúdo em um commit ou ref específico
func ContentAt(repo, path, ref string) ([]byte, error) {
	token := ResolveToken()
	if token == "" {
		out, err := GH(30*time.Second, "api", fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", Owner, repo, path, ref), "--jq", ".content")
		if err != nil {
			return nil, err
		}
		return decodeBase64(string(out))
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s?ref=%s", contentURL(repo, path), ref), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "bot-telemetry-go")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api status %d", resp.StatusCode)
	}

	var res struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return decodeBase64(res.Content)
}

// List lista os arquivos de um diretório
func List(repo, path string) ([]string, error) {
	token := ResolveToken()
	if token == "" {
		out, err := GH(30*time.Second, "api", fmt.Sprintf("repos/%s/%s/contents/%s", Owner, repo, path), "--jq", ".[].name")
		if err != nil {
			return nil, err
		}
		var names []string
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			line = strings.Trim(line, `"`)
			if line != "" {
				names = append(names, line)
			}
		}
		return names, nil
	}

	req, err := http.NewRequest("GET", contentURL(repo, path), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "bot-telemetry-go")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api status %d", resp.StatusCode)
	}

	var items []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	names := make([]string, len(items))
	for i, it := range items {
		names[i] = it.Name
	}
	return names, nil
}

// Put envia ou atualiza um arquivo no GitHub via HTTP direto (Zero subprocessos, Zero arquivos temporários no /tmp)
func Put(repo, path, message string, content []byte, sha string) error {
	token := ResolveToken()
	if token == "" {
		// Fallback para CLI
		payload := map[string]string{
			"message": message,
			"content": base64.StdEncoding.EncodeToString(content),
		}
		if sha != "" {
			payload["sha"] = sha
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		tmp, err := os.CreateTemp("", "ghput-*.json")
		if err != nil {
			return err
		}
		name := tmp.Name()
		defer os.Remove(name)
		if _, err := tmp.Write(data); err != nil {
			tmp.Close()
			return err
		}
		if err := tmp.Close(); err != nil {
			return err
		}
		_, err = GH(30*time.Second, "api", "-X", "PUT", fmt.Sprintf("repos/%s/%s/contents/%s", Owner, repo, path), "--input", name)
		return err
	}

	payload := map[string]string{
		"message": message,
		"content": base64.StdEncoding.EncodeToString(content),
	}
	if sha != "" {
		payload["sha"] = sha
	}
	bodyData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal put payload: %w", err)
	}

	req, err := http.NewRequest("PUT", contentURL(repo, path), bytes.NewReader(bodyData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "bot-telemetry-go")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http put error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		errResp, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api put status %d: %s", resp.StatusCode, string(errResp))
	}
	return nil
}

// API executa chamada genérica da API (fallback)
func API(timeout time.Duration, args ...string) ([]byte, error) {
	return GH(timeout, append([]string{"api"}, args...)...)
}

func decodeBase64(s string) ([]byte, error) {
	s = strings.Join(strings.Fields(s), "")
	if s == "" {
		return nil, fmt.Errorf("conteudo base64 vazio")
	}
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		raw, err2 := base64.RawStdEncoding.DecodeString(s)
		if err2 != nil {
			return nil, err
		}
		return raw, nil
	}
	return raw, nil
}
