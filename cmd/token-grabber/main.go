package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

func main() {
	userFilter := flag.String("user", os.Getenv("BOT_USER"), "Nome do usuário para filtrar token")
	vpsSync := flag.Bool("sync", false, "Sincronizar token com a VPS após captura")
	vpsHost := flag.String("host", os.Getenv("VPS_HOST"), "IP da VPS")
	vpsKey := flag.String("key", os.Getenv("VPS_KEY"), "Caminho da chave SSH")
	flag.Parse()
	_ = userFilter

	fmt.Println("==================================================")
	fmt.Println("  🔍 PokeTrue Token Grabber - Memória do Navegador")
	fmt.Println("==================================================")

	targetPattern := []byte("eyJpZCI")
	jwtRegex := regexp.MustCompile(`eyJpZCI[A-Za-z0-9_\-]{15,}\.[A-Za-z0-9_\-]{15,}\.[A-Za-z0-9_\-]+`)

	snap, _, _ := procCreateToolhelp.Call(uintptr(TH32CS_SNAPPROCESS), 0)
	if snap == 0 || snap == uintptr(syscall.InvalidHandle) {
		fmt.Println("❌ Falha ao criar snapshot do sistema operacional")
		os.Exit(1)
	}
	defer procCloseHandle.Call(snap)

	var entry PROCESSENTRY32W
	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32First.Call(snap, uintptr(unsafe.Pointer(&entry)))
	var browserPids []uint32

	for ret != 0 {
		name := syscall.UTF16ToString(entry.ExeFile[:])
		if name == "brave.exe" || name == "msedge.exe" || name == "chrome.exe" || name == "opera.exe" {
			browserPids = append(browserPids, entry.ProcessID)
		}
		ret, _, _ = procProcess32Next.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}

	fmt.Printf("🔎 Escaneando %d processos de navegadores...\n", len(browserPids))

	var latestToken string
	found := make(map[string]bool)

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
								if !found[tok] {
									found[tok] = true
									latestToken = tok
									fmt.Printf("✅ Token capturado [PID %d]: %s...%s\n", pid, tok[:20], tok[len(tok)-15:])
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

	if latestToken == "" {
		fmt.Println("❌ Nenhum token encontrado na memória dos navegadores.")
		fmt.Println("💡 Dica: Verifique se o jogo PokeTrue está aberto e logado em uma aba do navegador.")
		os.Exit(1)
	}

	// Atualizar accounts.json local
	accountsPath := "accounts.json"
	raw, err := os.ReadFile(accountsPath)
	if err == nil {
		var accounts []map[string]any
		if err := json.Unmarshal(raw, &accounts); err == nil && len(accounts) > 0 {
			if creds, ok := accounts[0]["credentials"].(map[string]any); ok {
				creds["token"] = latestToken
				updated, _ := json.MarshalIndent(accounts, "", "  ")
				_ = os.WriteFile(accountsPath, updated, 0644)
				fmt.Println("💾 accounts.json local atualizado com sucesso!")
			}
		}
	}

	// Sincronizar com VPS se solicitado
	if *vpsSync {
		fmt.Printf("🚀 Enviando para VPS (%s)...\n", *vpsHost)
		scpCmd := exec.Command("scp", "-i", *vpsKey, "-o", "StrictHostKeyChecking=no", accountsPath, fmt.Sprintf("ubuntu@%s:/home/ubuntu/poketrue-bot/accounts.json", *vpsHost))
		if out, err := scpCmd.CombinedOutput(); err != nil {
			fmt.Printf("⚠️ Erro no SCP: %v (%s)\n", err, string(out))
		} else {
			fmt.Println("📡 Reiniciando serviço na VPS...")
			sshCmd := exec.Command("ssh", "-i", *vpsKey, "-o", "StrictHostKeyChecking=no", fmt.Sprintf("ubuntu@%s", *vpsHost), "sudo systemctl restart poketrue.service")
			_ = sshCmd.Run()
			time.Sleep(1 * time.Second)
			fmt.Println("🎉 Bot na VPS atualizado e rodando com o novo token!")
		}
	}
}
