package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCapture(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ts() string {
	return time.Now().Format("20060102-150405")
}

func curlGet(url string) string {
	out, _ := runCapture("curl", "-s", "-m", "10", url)
	return out
}

func leakAudit(url string) {
	body := curlGet(url)
	lower := strings.ToLower(body)
	if strings.Contains(lower, "kbelludoo") || strings.Contains(lower, "lohmoreira") {
		fmt.Println("VAZAMENTO AINDA PRESENTE")
	} else {
		fmt.Println("LIMPO")
	}
}

func cmdHunteraWeb() {
	stamp := ts()
	run("sed", "-i", `s/\r$//`, "/tmp/web-server.stealth.js")
	run("cp", "/home/ubuntu/huntera-cli/web-server.js", "/home/ubuntu/huntera-cli/web-server.js.bak-"+stamp)
	run("cp", "/tmp/web-server.stealth.js", "/home/ubuntu/huntera-cli/web-server.js")
	fmt.Println("backup: web-server.js.bak-" + stamp)
	run("tmux", "kill-session", "-t", "huntera-web")
	time.Sleep(1 * time.Second)
	run("tmux", "new-session", "-d", "-s", "huntera-web", "cd /home/ubuntu/huntera-cli && node web-server.js")
	time.Sleep(2 * time.Second)
	fmt.Println("== teste API local ==")
	out := curlGet("http://127.0.0.1:3000/api/status")
	fmt.Println(out)
	fmt.Println("== teste via túnel público ==")
	out = curlGet("https://ron-flexible-respective-caring.trycloudflare.com/api/status")
	fmt.Println(out)
	fmt.Println("== leak audit via túnel ==")
	leakAudit("https://ron-flexible-respective-caring.trycloudflare.com/api/status")
}

func cmdHunteraGo() {
	stamp := ts()
	run("cp", "/home/ubuntu/huntera-go/huntera-bot", "/home/ubuntu/huntera-go/huntera-bot.bak-"+stamp)
	run("cp", "/home/ubuntu/huntera-cli/go/pkg/server/server.go", "/home/ubuntu/huntera-cli/go/pkg/server/server.go.bak-"+stamp)
	run("cp", "/tmp/server.go.stealth", "/home/ubuntu/huntera-cli/go/pkg/server/server.go")
	run("sed", "-i", `s/\r$//`, "/home/ubuntu/huntera-cli/go/pkg/server/server.go")
	fmt.Println("backups: huntera-bot.bak-" + stamp + " e server.go.bak-" + stamp)
	run("tmux", "kill-session", "-t", "huntera-go")
	time.Sleep(1 * time.Second)
	run("pkill", "-x", "huntera-bot")
	time.Sleep(2 * time.Second)
	run("pkill", "-9", "-x", "huntera-bot")
	time.Sleep(1 * time.Second)
	run("cp", "/tmp/huntera-bot-stealth", "/home/ubuntu/huntera-go/huntera-bot")
	run("chmod", "+x", "/home/ubuntu/huntera-go/huntera-bot")
	run("bash", "/home/ubuntu/huntera-go/start.sh")
	time.Sleep(8 * time.Second)
	fmt.Println("== teste API local ==")
	out := curlGet("http://127.0.0.1:3000/api/status")
	fmt.Println(out)
	fmt.Println("== teste via túnel público ==")
	out = curlGet("https://ron-flexible-respective-caring.trycloudflare.com/api/status")
	fmt.Println(out)
	fmt.Println("== leak audit via túnel ==")
	leakAudit("https://ron-flexible-respective-caring.trycloudflare.com/api/status")
}

func cmdBaiak() {
	stamp := ts()
	run("cp", "/home/ubuntu/baiak-bot-ts/src/term/fleet_server.ts", "/home/ubuntu/baiak-bot-ts/src/term/fleet_server.ts.bak-"+stamp)
	run("cp", "/tmp/fleet_server.stealth.ts", "/home/ubuntu/baiak-bot-ts/src/term/fleet_server.ts")
	run("sed", "-i", `s/\r$//`, "/home/ubuntu/baiak-bot-ts/src/term/fleet_server.ts")
	fmt.Println("backup: fleet_server.ts.bak-" + stamp)
	fmt.Println("== como o fleet server roda ==")
	run("systemctl", "--user", "list-units", "--type=service")
	run("systemctl", "list-units", "--type=service")
	run("tmux", "ls")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("uso: deploy <huntera-web|huntera-go|baiak>")
		fmt.Println("  huntera-web - aplica web-server.stealth.js e reinicia via tmux")
		fmt.Println("  huntera-go  - aplica binário stealth e reinicia via start.sh")
		fmt.Println("  baiak       - aplica fleet_server.stealth.ts")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "huntera-web":
		cmdHunteraWeb()
	case "huntera-go":
		cmdHunteraGo()
	case "baiak":
		cmdBaiak()
	default:
		fmt.Printf("comando desconhecido: %s\n", os.Args[1])
		os.Exit(1)
	}
}
