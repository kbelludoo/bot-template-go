package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess    = kernel32.NewProc("OpenProcess")
	procCloseHandle    = kernel32.NewProc("CloseHandle")
	procVirtualQueryEx = kernel32.NewProc("VirtualQueryEx")
	procReadProcMem    = kernel32.NewProc("ReadProcessMemory")
	procCreateToolhelp = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First = kernel32.NewProc("Process32FirstW")
	procProcess32Next  = kernel32.NewProc("Process32NextW")
)

const (
	PROCESS_QUERY_INFORMATION = 0x0400
	PROCESS_VM_READ           = 0x0010
	MEM_COMMIT                = 0x1000
	PAGE_NOACCESS             = 0x01
	PAGE_GUARD                = 0x100
	TH32CS_SNAPPROCESS        = 0x00000002
)

type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	PartitionId       uint16
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

type PROCESSENTRY32W struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

type JWTPayload struct {
	ID     int    `json:"id"`
	User   string `json:"user"`
	PJ     int    `json:"pj"`
	PJNome string `json:"pjNome"`
	Exp    int64  `json:"exp"`
}

func parseJWT(token string) *JWTPayload {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payloadBytes, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil
		}
	}
	var p JWTPayload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		return nil
	}
	return &p
}

func scanBrowserMemory() []string {
	targetPattern := []byte("eyJpZCI")
	jwtRegex := regexp.MustCompile(`eyJpZCI[A-Za-z0-9_\-]{15,}\.[A-Za-z0-9_\-]{15,}\.[A-Za-z0-9_\-]+`)

	snap, _, _ := procCreateToolhelp.Call(uintptr(TH32CS_SNAPPROCESS), 0)
	if snap == 0 || snap == uintptr(syscall.InvalidHandle) {
		return nil
	}
	defer procCloseHandle.Call(snap)

	var entry PROCESSENTRY32W
	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32First.Call(snap, uintptr(unsafe.Pointer(&entry)))
	var browserPids []uint32

	for ret != 0 {
		name := strings.ToLower(syscall.UTF16ToString(entry.ExeFile[:]))
		if strings.Contains(name, "brave") || strings.Contains(name, "msedge") || strings.Contains(name, "chrome") || strings.Contains(name, "opera") {
			browserPids = append(browserPids, entry.ProcessID)
		}
		ret, _, _ = procProcess32Next.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}

	foundMap := make(map[string]bool)
	var tokens []string

	for _, pid := range browserPids {
		hProc, _, _ := procOpenProcess.Call(PROCESS_QUERY_INFORMATION|PROCESS_VM_READ, 0, uintptr(pid))
		if hProc == 0 {
			continue
		}

		var addr uintptr = 0
		var mbi MEMORY_BASIC_INFORMATION

		for {
			r, _, _ := procVirtualQueryEx.Call(hProc, addr, uintptr(unsafe.Pointer(&mbi)), unsafe.Sizeof(mbi))
			if r == 0 {
				break
			}

			if mbi.State == MEM_COMMIT && (mbi.Protect&PAGE_GUARD) == 0 && (mbi.Protect&PAGE_NOACCESS) == 0 {
				chunkSize := uintptr(4 * 1024 * 1024)
				for offset := uintptr(0); offset < mbi.RegionSize; offset += chunkSize {
					toRead := chunkSize
					if offset+toRead > mbi.RegionSize {
						toRead = mbi.RegionSize - offset
					}

					buf := make([]byte, toRead)
					var bytesRead uintptr
					procReadProcMem.Call(hProc, mbi.BaseAddress+offset, uintptr(unsafe.Pointer(&buf[0])), toRead, uintptr(unsafe.Pointer(&bytesRead)))
					if bytesRead > 0 {
						buf = buf[:bytesRead]
						if bytes.Contains(buf, targetPattern) {
							matches := jwtRegex.FindAll(buf, -1)
							for _, m := range matches {
								tok := string(m)
								if !foundMap[tok] {
									p := parseJWT(tok)
									if p != nil && p.User != "" {
										foundMap[tok] = true
										tokens = append(tokens, tok)
									}
								}
							}
						}
					}
				}
			}

			addr = mbi.BaseAddress + mbi.RegionSize
			if addr == 0 {
				break
			}
		}

		procCloseHandle.Call(hProc)
	}

	return tokens
}

func main() {
	defaultKey := filepath.Join(os.Getenv("USERPROFILE"), ".ssh", "ssh-key-2026-10-06.key")
	accFlag := flag.Int("acc", 2, "Número da conta a atualizar (1 ou 2)")
	watchFlag := flag.Bool("watch", false, "Modo contínuo: monitora a memória até o login acontecer")
	syncFlag := flag.Bool("sync", true, "Sincronizar com VPS após captura")
	vpsHost := flag.String("host", "129.153.157.17", "IP da VPS")
	vpsKey := flag.String("key", defaultKey, "Caminho da chave SSH")
	manualToken := flag.String("token", "", "Token manual se já copiado pelo usuário")
	flag.Parse()

	fmt.Println("==================================================")
	fmt.Println("  🔍 PokeTrue Token Grabber & Multi-Account Sync  ")
	fmt.Println("==================================================")

	var targetToken string

	if *manualToken != "" {
		targetToken = strings.TrimSpace(*manualToken)
		p := parseJWT(targetToken)
		if p != nil {
			fmt.Printf("✅ Token manual fornecido para usuário: %s (PJ: %s | ID: %d)\n", p.User, p.PJNome, p.ID)
		}
	} else {
		for {
			fmt.Println("🔎 Escaneando memória dos navegadores (Brave / Chrome / Edge)...")
			tokens := scanBrowserMemory()

			if len(tokens) > 0 {
				fmt.Printf("🎯 %d token(s) válidos encontrados na RAM:\n", len(tokens))
				for i, t := range tokens {
					p := parseJWT(t)
					user := "desconhecido"
					if p != nil {
						user = p.User
					}
					fmt.Printf("  [%d] Usuário: %-15s | Token: %s...%s\n", i+1, user, t[:15], t[len(t)-10:])
					if *accFlag == 2 && user != "kbelludoo" {
						targetToken = t
					} else if *accFlag == 1 && user == "kbelludoo" {
						targetToken = t
					}
				}

				if targetToken == "" && len(tokens) > 0 {
					targetToken = tokens[len(tokens)-1]
				}

				if targetToken != "" {
					break
				}
			}

			if !*watchFlag {
				break
			}
			fmt.Println("⏳ Aguardando login no navegador... (tentando novamente em 3 segundos)")
			time.Sleep(3 * time.Second)
		}
	}

	if targetToken == "" {
		fmt.Println("❌ Nenhum novo token encontrado na memória dos navegadores.")
		fmt.Println("💡 Dica: Abra https://poketrue.net/play no navegador, faça login na Conta 2 e rode novamente.")
		os.Exit(1)
	}

	// Atualizar accounts.json local
	accountsPath := "accounts.json"
	raw, err := os.ReadFile(accountsPath)
	if err != nil {
		fmt.Printf("❌ Erro ao ler accounts.json: %v\n", err)
		os.Exit(1)
	}

	var accounts []map[string]any
	if err := json.Unmarshal(raw, &accounts); err != nil {
		fmt.Printf("❌ Erro ao interpretar accounts.json: %v\n", err)
		os.Exit(1)
	}

	accIdx := *accFlag - 1
	if accIdx < 0 || accIdx >= len(accounts) {
		accIdx = len(accounts) - 1
	}

	if creds, ok := accounts[accIdx]["credentials"].(map[string]any); ok {
		creds["token"] = targetToken
		updated, _ := json.MarshalIndent(accounts, "", "  ")
		_ = os.WriteFile(accountsPath, updated, 0644)
		p := parseJWT(targetToken)
		userName := "desconhecido"
		if p != nil {
			userName = p.User
		}
		fmt.Printf("💾 accounts.json local atualizado para Conta #%d (%s)!\n", accIdx+1, userName)
	}

	// Sincronizar com VPS
	if *syncFlag {
		fmt.Printf("🚀 Enviando accounts.json atualizado para a VPS (%s)...\n", *vpsHost)
		scpCmd := exec.Command("scp", "-i", *vpsKey, "-o", "StrictHostKeyChecking=no", accountsPath, fmt.Sprintf("ubuntu@%s:/home/ubuntu/poketrue-bot/accounts.json", *vpsHost))
		if out, err := scpCmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️ Erro no SCP: %v (%s)\n", err, string(out))
		} else {
			fmt.Println("📡 Reiniciando serviço poketrue.service na VPS...")
			sshCmd := exec.Command("ssh", "-i", *vpsKey, "-o", "StrictHostKeyChecking=no", fmt.Sprintf("ubuntu@%s", *vpsHost), "sudo systemctl restart poketrue.service")
			_ = sshCmd.Run()
			time.Sleep(1 * time.Second)
			fmt.Println("🎉 poketrue.service reiniciado na VPS com Sucesso!")
		}
	}
}
